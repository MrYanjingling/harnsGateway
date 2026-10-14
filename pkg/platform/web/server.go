package web

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/gin-gonic/gin"
	"harns-platform/cmd/platform/config"
	"harns-platform/cmd/platform/options"
	"harns-platform/pkg/apis/response"
	"harns-platform/pkg/generic"
	"harns-platform/pkg/logging"
	"net/http"
)

type Server struct {
	*generic.Server
	*config.Config
}

func NewServer(router *gin.Engine, o *options.Options, config *config.Config) (*Server, error) {
	allowMethods := []string{http.MethodPost, http.MethodGet, http.MethodDelete, http.MethodPut, http.MethodPatch}

	s := &generic.Server{
		Router:  router,
		Port:    o.Port,
		Methods: allowMethods,
	}

	server := &Server{
		Server: s,
		Config: config,
	}

	server.InstallHandlers()

	return server, nil
}

func (s *Server) InstallHandlers() {
	v1 := s.Router.Group("/api/v1")
	v1.GET("/health", func(c *gin.Context) {
		response.WriteOK(c, gin.H{"status": "ok"})
	})
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
			Addr:      fmt.Sprintf(":%s", s.Port),
			Handler:   s.Router,
			TLSConfig: c,
		}
		go func() {
			logging.Error(srv.ListenAndServeTLS("", ""))
		}()
	} else {
		srv = &http.Server{
			Addr:    fmt.Sprintf(":%s", s.Port),
			Handler: s.Router,
		}
		go func() {
			logging.Error(srv.ListenAndServe())
		}()
	}

	return func(ctx context.Context) {
		srv.SetKeepAlivesEnabled(false)
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
