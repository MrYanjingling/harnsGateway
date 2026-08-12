package dataweb

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"harnsgateway/cmd/data/config"
	"harnsgateway/cmd/data/options"
	"harnsgateway/pkg/data"
	"harnsgateway/pkg/generic"
)

type Server struct {
	*generic.Server
	*config.Config
}

func NewServer(router *gin.Engine, o *options.Options, c *config.Config) (*Server, error) {
	allowMethods := []string{http.MethodPost, http.MethodGet, http.MethodDelete, http.MethodPut, http.MethodPatch}

	s := &generic.Server{
		Router:  router,
		Port:    o.Port,
		Methods: allowMethods,
	}

	server := &Server{
		Server: s,
		Config: c,
	}

	server.InstallHandlers()

	return server, nil
}

func (s *Server) InstallHandlers() {
	v1 := s.Router.Group("/api/v1")
	data.InstallHandler(v1, s.Config.DataManager)
}

func (s *Server) Serve() (func(ctx context.Context), error) {
	var srv *http.Server
	if len(s.Config.CertFile) != 0 && len(s.Config.KeyFile) != 0 {
		x509KeyPair, err := tls.LoadX509KeyPair(s.Config.CertFile, s.Config.KeyFile)
		if err != nil {
			return nil, err
		}
		tlsCfg := &tls.Config{
			Certificates: []tls.Certificate{x509KeyPair},
		}

		srv = &http.Server{
			Addr:      fmt.Sprintf(":%s", s.Port),
			Handler:   s.Router,
			TLSConfig: tlsCfg,
		}
		go func() {
			klog.Error(srv.ListenAndServeTLS("", ""))
		}()
	} else {
		srv = &http.Server{
			Addr:    fmt.Sprintf(":%s", s.Port),
			Handler: s.Router,
		}
		go func() {
			klog.Error(srv.ListenAndServe())
		}()
	}

	return func(ctx context.Context) {
		srv.SetKeepAlivesEnabled(false)
		if s.Config.DB != nil {
			if err := s.Config.DB.Close(); err != nil {
				klog.ErrorS(err, "Failed to close MySQL connection")
			}
		}
		if err := srv.Shutdown(ctx); err != nil {
			klog.ErrorS(err, "Failed to shutdown HTTP server")
		}
	}, nil
}
