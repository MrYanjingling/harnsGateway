package main

import (
	"harnsgateway/cmd/cim/app"
	"k8s.io/component-base/logs"
	_ "k8s.io/component-base/logs/json/register"
	"os"
)

func main() {

	cmd := app.NewCimCmd()
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
