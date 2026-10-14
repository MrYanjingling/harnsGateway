package config

import (
	"harns-platform/pkg/gateway"
	"harns-platform/pkg/gateway/device"
)

type Config struct {
	DeviceMgr  *device.Manager
	GatewayMgr *gateway.Manager
	CertFile   string
	KeyFile    string
}
