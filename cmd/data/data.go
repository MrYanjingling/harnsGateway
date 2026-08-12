package main

import (
	"harnsgateway/cmd/data/app"
	"os"

	"k8s.io/component-base/logs"
	_ "k8s.io/component-base/logs/json/register"
)

func main() {
	cmd := app.NewDataCmd()
	logs.InitLogs()
	defer logs.FlushLogs()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
