package main

import (
	"harns-platform/cmd/gateway/app"
	"harns-platform/pkg/logging"
	"os"
)

func main() {
	cmd := app.NewGatewayCmd()
	defer logging.Flush()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
