package app

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"harnsgateway/cmd/iotsimulator/options"
	baseoptions "harnsgateway/pkg/generic/options"
	"harnsgateway/pkg/version"
	"harnsgateway/pkg/version/verflag"
	utilserrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	ComponentIotSimulator = "iot-simulator"
)

func NewIotSimulatorCmd() *cobra.Command {
	cleanFlagSet := pflag.NewFlagSet(ComponentIotSimulator, pflag.ContinueOnError)
	o := options.NewDefaultOptions()
	cmd := &cobra.Command{
		Use:                ComponentIotSimulator,
		Long:               `The harns gateway manages the device, collect and control.`,
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

	_, err := o.Config(stopCh)
	if err != nil {
		return err
	}

	// Graceful shutdown
	// Wait for interrupt signal to gracefully shutdown the server
	exitCh := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(exitCh, syscall.SIGINT, syscall.SIGTERM)
	<-exitCh
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// exit(ctx)
	close(stopCh)

	return nil
}
