package config

import (
	"harns-platform/pkg/gateway"
	"harns-platform/pkg/gateway/device"
	"harns-platform/pkg/generic"
)

type Config struct {
	DeviceMgr    *device.Manager
	GatewayMgr   *gateway.Manager
	CertFile     string
	KeyFile      string
	MaxBodyBytes int64
	CORS         generic.CORSConfig
}
