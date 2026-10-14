package app

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"harns-platform/cmd/platform/options"
	"harns-platform/pkg/generic"
	baseoptions "harns-platform/pkg/generic/options"
	"harns-platform/pkg/logging"
	"harns-platform/pkg/platform/web"
	"harns-platform/pkg/version"
	"harns-platform/pkg/version/verflag"
	utilserrors "k8s.io/apimachinery/pkg/util/errors"
	"os"
	"os/signal"
	"syscall"
)

const (
	ComponentPlatform = "harns-platform"
)

func NewPlatformCmd() *cobra.Command {
	cleanFlagSet := pflag.NewFlagSet(ComponentPlatform, pflag.ContinueOnError)
	o := options.NewDefaultOptions()
	cmd := &cobra.Command{
		Use:                ComponentPlatform,
		Long:               `The harns platform is the IoT basic platform service.`,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// initial flag parse, since we disable cobra's flag parsing
			if err := cleanFlagSet.Parse(args); err != nil {
				logging.ErrorS(err, "Failed to parse flag")
				_ = cmd.Usage()
				os.Exit(1)
			}

			// check if there are non-flag arguments in the command line
			cmds := cleanFlagSet.Args()
			if len(cmds) > 0 {
				logging.ErrorS(nil, "Unknown command", "command", cmds[0])
				_ = cmd.Usage()
				os.Exit(1)
			}

			verflag.PrintAndExitIfRequested()
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
				return utilserrors.NewAggregate(errs)
			}

			// To help debugging, immediately log version
			logging.Infof("Version: %+v", version.Get())
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

	c, err := o.Config(stopCh)
	if err != nil {
		return err
	}

	server, err := web.NewServer(generic.Default(), o, c)
	if err != nil {
		return err
	}

	exit, err := server.Serve()
	if err != nil {
		return err
	}
	logging.V(1).InfoS("Server started", "port", o.Port)

	// Graceful shutdown
	// Wait for interrupt signal to gracefully shutdown the server
	exitCh := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(exitCh, syscall.SIGINT, syscall.SIGTERM)
	<-exitCh
	ctx, cancel := context.WithTimeout(context.Background(), o.Wait.Duration())
	defer cancel()

	exit(ctx)
	close(stopCh)

	return nil
}
