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
	"harns-platform/pkg/audit"
	"harns-platform/pkg/generic"
	"harns-platform/pkg/logging"
	"harns-platform/pkg/platform/auditapi"
	"harns-platform/pkg/platform/auth"
	"harns-platform/pkg/platform/httpx"
	"harns-platform/pkg/platform/identity"
	"harns-platform/pkg/telemetry"
)

type Server struct {
	*generic.Server
	*config.Config
	audit       *audit.Module
	maintenance *audit.Maintenance
}

func NewServer(router *gin.Engine, o *options.Options, config *config.Config) (*Server, error) {
	if config.Telemetry.Enabled {
		router.Use(telemetry.Middleware(config.Telemetry.ServiceName))
	}
	router.Use(generic.CORS(config.CORS))
	router.Use(httpx.Recovery())
	router.HandleMethodNotAllowed = true
	router.NoRoute(func(c *gin.Context) {
		response.WriteError(c, http.StatusNotFound, response.ErrNotFound())
	})
	router.NoMethod(func(c *gin.Context) {
		response.WriteError(c, http.StatusMethodNotAllowed, response.ErrMethodNotAllowed())
	})

	s := &generic.Server{
		Router: router,
		Port:   o.Port,
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
	v1.Use(generic.BodyLimit(s.Config.MaxBodyBytes))
	v1.GET("/health", func(c *gin.Context) {
		response.WriteOK(c, gin.H{"status": "ok"})
	})
	v1.GET("/ready", readiness(s.Config))
	installOpenAPI(s.Router, v1)

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
		s.audit = audit.NewModule(s.Config.DB, s.Config.Audit, audit.AppIOT,
			audit.WithMetrics(auditapi.PlatformMetrics{}))
		auditMW = auditapi.Middleware(s.audit, s.Config.Audit)
	}
	s.maintenance = audit.NewMaintenance(s.Config.DB, s.Config.Audit)
	reader := audit.NewReader(s.Config.DB, "")

	authGroup := v1.Group("/auth")
	authGroup.POST("/login", auditMW, httpx.WithOperation(httpx.OpLogin, "用户登录"), authSvc.Login)
	authGroup.POST("/logout", auditMW, httpx.WithOperation(httpx.OpLogout, "用户退出"), authSvc.Logout)
	authGroup.POST("/verificationCode", authSvc.VerificationCode)
	authGroup.GET("/publicKey", authSvc.PublicKey)
	authGroup.GET("/index", authSvc.Auth(), authSvc.Index)

	iot := v1.Group("/iot", authSvc.Auth(), auditMW)
	identity.InstallHandler(iot, identitySvc)
	iot.GET("/auditLogs", httpx.WithOperation(httpx.OpQuery, "查询审计日志"), auditapi.List(reader))
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
		// Stop accepting new requests and drain in-flight ones first.
		if err := srv.Shutdown(ctx); err != nil {
			logging.Error(err)
		}
		// Audit writer may still be persisting logs, so close it before the DB.
		if s.audit != nil {
			_ = s.audit.Close(ctx)
		}
		if s.maintenance != nil {
			s.maintenance.Close(ctx)
		}
		if s.Config.TelemetryShutdown != nil {
			tctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := s.Config.TelemetryShutdown(tctx); err != nil {
				logging.Error(err)
			}
			cancel()
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
	}, nil
}
