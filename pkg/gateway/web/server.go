package web

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"harns-platform/cmd/gateway/config"
	"harns-platform/cmd/gateway/options"
	"harns-platform/pkg/apis/response"
	"harns-platform/pkg/gateway"
	"harns-platform/pkg/gateway/device"
	"harns-platform/pkg/generic"
	"harns-platform/pkg/logging"
	"net/http"
)

type Server struct {
	*generic.Server
	*config.Config
}

func NewServer(router *gin.Engine, o *options.Options, config *config.Config) (*Server, error) {
	router.Use(generic.CORS(config.CORS))
	router.Use(gin.Recovery())
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

	server.InstallHandlers()

	return server, nil

}

func (s *Server) InstallHandlers() {
	v1 := s.Router.Group("/api/v1")
	v1.Use(generic.BodyLimit(s.Config.MaxBodyBytes))
	device.InstallHandler(v1, s.Config.DeviceMgr)
	gateway.InstallHandler(v1, s.Config.GatewayMgr)
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
		// Drain HTTP requests before tearing down the device manager.
		if err := srv.Shutdown(ctx); err != nil {
			logging.Error(err)
		}
		if err := s.Config.DeviceMgr.Shutdown(ctx); err != nil {
			logging.Error(err)
		}
	}, nil
}
