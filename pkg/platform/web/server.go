package web

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"harns-platform/cmd/platform/config"
	"harns-platform/cmd/platform/options"
	"harns-platform/pkg/apis/response"
	"harns-platform/pkg/generic"
	"harns-platform/pkg/logging"
	"harns-platform/pkg/platform/audit"
	"harns-platform/pkg/platform/auth"
	"harns-platform/pkg/platform/httpx"
	"harns-platform/pkg/platform/identity"
)

type Server struct {
	*generic.Server
	*config.Config
	audit *audit.Service
}

func NewServer(router *gin.Engine, o *options.Options, config *config.Config) (*Server, error) {
	allowMethods := []string{http.MethodPost, http.MethodGet, http.MethodDelete, http.MethodPut, http.MethodPatch}

	router.Use(httpx.CORS(config.CORS))
	router.Use(httpx.Recovery())
	router.HandleMethodNotAllowed = true
	router.NoRoute(func(c *gin.Context) {
		response.WriteError(c, http.StatusNotFound, response.ErrNotFound())
	})
	router.NoMethod(func(c *gin.Context) {
		response.WriteError(c, http.StatusMethodNotAllowed, response.ErrMethodNotAllowed())
	})

	s := &generic.Server{
		Router:  router,
		Port:    o.Port,
		Methods: allowMethods,
	}

	server := &Server{
		Server: s,
		Config: config,
	}

	if err := server.InstallHandlers(); err != nil {
		return nil, err
	}

	return server, nil
}

func (s *Server) InstallHandlers() error {
	v1 := s.Router.Group("/api/v1")
	v1.Use(httpx.Trace(), httpx.BodyLimit(maxRequestBodyBytes))
	v1.GET("/health", func(c *gin.Context) {
		response.WriteOK(c, gin.H{"status": "ok"})
	})
	v1.GET("/ready", readiness(s.Config))

	if s.Config.DB == nil {
		logging.Warnf("database not configured, identity endpoints disabled")
		return nil
	}

	identitySvc := identity.NewService(identity.NewStore(s.Config.DB))
	authSvc, err := auth.NewService(identitySvc, s.Config.Redis, s.Config.MinIO, s.Config.MinIOBucket, s.Config.Security)
	if err != nil {
		return fmt.Errorf("init auth service failed: %w", err)
	}
	identitySvc.SetSessionInvalidator(authSvc)

	var auditMW gin.HandlerFunc = func(c *gin.Context) { c.Next() }
	if s.Config.Audit.Enabled {
		s.audit = audit.NewService(s.Config.DB, s.Config.Audit)
		auditMW = s.audit.Middleware()
	}

	authGroup := v1.Group("/auth")
	authGroup.POST("/login", auditMW, httpx.WithOperation(httpx.OpLogin, "用户登录"), authSvc.Login)
	authGroup.POST("/logout", auditMW, httpx.WithOperation(httpx.OpLogout, "用户退出"), authSvc.Logout)
	authGroup.POST("/verificationCode", authSvc.VerificationCode)
	authGroup.GET("/publicKey", authSvc.PublicKey)
	authGroup.GET("/index", authSvc.Auth(), authSvc.Index)

	iot := v1.Group("/iot", authSvc.Auth(), auditMW)
	identity.InstallHandler(iot, identitySvc)
	if s.audit != nil {
		iot.GET("/audit-logs", httpx.WithOperation(httpx.OpQuery, "查询审计日志"), s.audit.List)
	}
	return nil
}

func (s *Server) Serve() (func(ctx context.Context), error) {
	var srv *http.Server
	if len(s.Config.CertFile) != 0 && len(s.Config.KeyFile) != 0 {
		x509KeyPair, err := tls.LoadX509KeyPair(s.Config.CertFile, s.Config.KeyFile)
		if err != nil {
			return nil, err
		}
		c := &tls.Config{
			Certificates: []tls.Certificate{x509KeyPair},
		}

		srv = &http.Server{
			Addr:              fmt.Sprintf(":%s", s.Port),
			Handler:           s.Router,
			TLSConfig:         c,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      60 * time.Second,
			IdleTimeout:       120 * time.Second,
			MaxHeaderBytes:    1 << 20,
		}
		go func() {
			logging.Error(srv.ListenAndServeTLS("", ""))
		}()
	} else {
		srv = &http.Server{
			Addr:              fmt.Sprintf(":%s", s.Port),
			Handler:           s.Router,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      60 * time.Second,
			IdleTimeout:       120 * time.Second,
			MaxHeaderBytes:    1 << 20,
		}
		go func() {
			logging.Error(srv.ListenAndServe())
		}()
	}

	return func(ctx context.Context) {
		srv.SetKeepAlivesEnabled(false)
		if s.audit != nil {
			s.audit.Close()
		}
		if s.Config.DB != nil {
			if err := s.Config.DB.Close(); err != nil {
				logging.Error(err)
			}
		}
		if s.Config.Redis != nil {
			if err := s.Config.Redis.Close(); err != nil {
				logging.Error(err)
			}
		}
		if err := srv.Shutdown(ctx); err != nil {
			logging.Error(err)
		}
	}, nil
}
