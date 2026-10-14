package main

import (
	"harns-platform/cmd/platform/app"
	"harns-platform/pkg/logging"
	"os"
)

func main() {
	cmd := app.NewPlatformCmd()
	defer logging.Flush()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
