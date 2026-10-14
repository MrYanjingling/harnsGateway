package main

import (
	"github.com/MrYanjingling/harns-platform/cmd/gateway/app"
	"github.com/MrYanjingling/harns-platform/pkg/logging"
	"os"
)

func main() {
	cmd := app.NewGatewayCmd()
	defer logging.Flush()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
