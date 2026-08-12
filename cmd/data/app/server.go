package app

import (
	"context"
	"harnsgateway/pkg/dataweb"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"

	"harnsgateway/cmd/data/options"
	"harnsgateway/pkg/generic"
	baseoptions "harnsgateway/pkg/generic/options"
	"harnsgateway/pkg/version"
	"harnsgateway/pkg/version/verflag"
)

const (
	ComponentData = "harns-data"
)

func NewDataCmd() *cobra.Command {
	cleanFlagSet := pflag.NewFlagSet(ComponentData, pflag.ContinueOnError)
	o := options.NewDefaultOptions()
	cmd := &cobra.Command{
		Use:                ComponentData,
		Long:               `The harns data service provides data access and storage capabilities.`,
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
				return errors.NewAggregate(errs)
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

	c, err := o.Config(stopCh)
	if err != nil {
		return err
	}

	server, err := dataweb.NewServer(generic.Default(), o, c)
	if err != nil {
		return err
	}

	exit, err := server.Serve()
	if err != nil {
		return err
	}
	klog.V(1).InfoS("Server started", "port", o.Port)

	// Graceful shutdown
	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, syscall.SIGINT, syscall.SIGTERM)
	<-exitCh

	ctx, cancel := context.WithTimeout(context.Background(), o.Wait)
	defer cancel()

	exit(ctx)
	close(stopCh)

	return nil
}
