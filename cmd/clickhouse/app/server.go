package app

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"harnsgateway/cmd/clickhouse/config"
	"harnsgateway/cmd/clickhouse/options"
	"harnsgateway/pkg/ck"
	"harnsgateway/pkg/generic"
	baseoptions "harnsgateway/pkg/generic/options"
	"harnsgateway/pkg/version"
	"harnsgateway/pkg/version/verflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	_ "k8s.io/component-base/logs/json/register"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	ComponentConsumer = "ck"
)

func NewCkCmd() *cobra.Command {
	cleanFlagSet := pflag.NewFlagSet(ComponentConsumer, pflag.ContinueOnError)
	o := options.NewDefaultOptions()
	cmd := &cobra.Command{
		Use:                ComponentConsumer,
		Long:               `The IoT Consumer is a daemon that consumes and handles the jobs of IoT.`,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// initial flag parse, since we disable cobra's flag parsing
			if err := cleanFlagSet.Parse(args); err != nil {
				klog.ErrorS(err, "Failed to parse flag")
				_ = cmd.Usage()
				os.Exit(1)
			}

			// check if there are non-flag arguments in the command line
			cmds := cleanFlagSet.Args()
			if len(cmds) > 0 {
				klog.ErrorS(nil, "Unknown command", "command", cmds[0])
				_ = cmd.Usage()
				os.Exit(1)
			}

			// short-circuit on help
			baseoptions.PrintHelpAndExitIfRequested(cmd, cleanFlagSet)

			// short-circuit on defaultconfig
			baseoptions.PrintDefaultConfigAndExitIfRequested(options.NewDefaultOptions(), cleanFlagSet)

			// short-circuit on verflag
			verflag.PrintAndExitIfRequested()

			if err := baseoptions.ParseAndApplyConfigFile(o, args); err != nil {
				return err
			}

			if errs := options.Validate(o); len(errs) != 0 {
				return utilerrors.NewAggregate(errs)
			}

			// To help debugging, immediately log version
			klog.Infof("Version: %+v", version.Get())
			return run(o)
		},
	}

	verflag.AddFlags(cleanFlagSet)
	o.AddFlags(cleanFlagSet)
	o.AddBaseFlags(cmd, cleanFlagSet)

	return cmd
}

func run(o *options.Options) error {
	stopCh := make(chan struct{})
	m, err := o.Config(stopCh)
	if err != nil {
		return err
	}

	server, err := NewServer(generic.Default(), o, m)
	if err != nil {
		return err
	}

	exit, err := server.Serve()
	if err != nil {
		return err
	}

	daemon, err := Daemon(m.CkMgr)
	if err != nil {
		return err
	}

	klog.V(1).InfoS("Server started", "port", o.Port)
	// Graceful shutdown
	// Wait for interrupt signal to gracefully shutdown the server
	exitCh := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(exitCh, syscall.SIGINT, syscall.SIGTERM)
	<-exitCh
	_, cancel := context.WithTimeout(context.Background(), o.Wait)
	defer cancel()
	daemon(context.Background())
	exit(context.Background())
	// exit(ctx)
	close(stopCh)

	return nil
}

func Daemon(manager *ck.Manager) (func(ctx context.Context), error) {
	cron := NewWithSeconds()
	if _, err := cron.AddFunc("* * * * *", func() {
		manager.Polling()
	}); err != nil {
		klog.V(2).InfoS("Failed to collect FMCS data", "err", err)
	}

	cron.Start()

	return func(ctx context.Context) {

	}, nil
}

func NewWithSeconds() *cron.Cron {
	secondParser := cron.NewParser(cron.Second | cron.Minute |
		cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)
	return cron.New(cron.WithParser(secondParser), cron.WithChain())
}

type Server struct {
	*generic.Server
	*config.Config
}

func NewServer(router *gin.Engine, o *options.Options, config *config.Config) (*Server, error) {
	allowMethods := []string{http.MethodPost, http.MethodGet, http.MethodDelete, http.MethodPut, http.MethodPatch}

	s := &generic.Server{
		Router:  router,
		Port:    "8812",
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
	_ = s.Router.Group("/api/v1")
	// data.InstallHandler(v1, s.Config.CkMgr)
}

func (s *Server) Serve() (func(ctx context.Context), error) {
	var srv *http.Server

	srv = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.Port),
		Handler: s.Router,
	}
	go func() {
		klog.Error(srv.ListenAndServe())
	}()

	return func(ctx context.Context) {
		srv.SetKeepAlivesEnabled(false)
		if err := s.Config.CkMgr.Shutdown(ctx); err != nil {
			klog.Error(err)
		}
		if err := srv.Shutdown(ctx); err != nil {
			klog.Error(err)
		}
	}, nil
}
