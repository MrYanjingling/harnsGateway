package config

import (
	"github.com/MrYanjingling/harns-platform/pkg/device"
	"github.com/MrYanjingling/harns-platform/pkg/gateway"
)

type Config struct {
	DeviceMgr  *device.Manager
	GatewayMgr *gateway.Manager
	CertFile   string
	KeyFile    string
}
