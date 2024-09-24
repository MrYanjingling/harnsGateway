package options

import (
	"github.com/go-redis/redis/v8"
	"github.com/spf13/pflag"
	"harnsgateway/cmd/gateway/config"
	"harnsgateway/pkg/device"
	"harnsgateway/pkg/gateway"
	"harnsgateway/pkg/generic"
	baseoptions "harnsgateway/pkg/generic/options"
	"harnsgateway/pkg/storage"
	"harnsgateway/pkg/ts"
	"time"
)

type Options struct {
	Port string        `json:"port"`
	Wait time.Duration `json:"graceful-timeout"`
	// MqttBrokerUrls []string      `json:"mqtt-broker-urls"`
	// MqttUsername   string        `json:"mqtt-username"`
	// MqttPassword   string        `json:"mqtt-password"`
	CertFile    string `json:"cert-file"`
	KeyFile     string `json:"key-file"`
	RedisUrl    string `json:"redis-url"`
	Password    string `json:"password"`
	DB          int    `json:"db"`
	Placeholder string `json:"placeholder"`
	TsUrl       string `json:"ts-url"`
	TsToken     string `json:"ts-token"`
	baseoptions.BaseOptions
	// logs.BaseOptions
}

const (
	_defaultPort          = "32200"
	_defaultWait          = 15 * time.Second
	_defaultRedisUrl      = "10.56.223.27:8596"
	_defaultRedisPassword = "Di@redis#1TScsG"
	_defaultRedisDB       = 0
	_defaultPlaceholder   = "@"
	_defaultTsUrl         = "http://10.56.223.27:8086"
	_defaultTsToken       = "Token c6jGYUCinwzeTWdeUh32"
)

// var (
// 	_defaultMqttBrokerUrls = []string{"tcp://127.0.0.1:1883"}
// )

func NewDefaultOptions() *Options {
	return &Options{
		Port:        _defaultPort,
		Wait:        _defaultWait,
		RedisUrl:    _defaultRedisUrl,
		Password:    _defaultRedisPassword,
		DB:          _defaultRedisDB,
		Placeholder: _defaultPlaceholder,
		TsUrl:       _defaultTsUrl,
		TsToken:     _defaultTsToken,
		BaseOptions: baseoptions.NewDefaultBaseOptions(),
		// CertFile:       "",
		// KeyFile:        "",
		// BaseOptions: logs.NewOptions(),
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	// refer to node port assignment https://rancher.com/docs/rancher/v2.x/en/installation/requirements/ports/#commonly-used-ports
	fs.StringVarP(&o.Port, "port", "P", o.Port, "Port exposed")
	fs.DurationVar(&o.Wait, "graceful-timeout", o.Wait, "The duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	fs.StringVarP(&o.RedisUrl, "redis-url", "", o.RedisUrl, "The redis server url")
	fs.StringVarP(&o.Password, "redis-password", "p", o.Password, "The redis password")
	fs.IntVarP(&o.DB, "redis-db", "d", o.DB, "The redis db")
	fs.StringVarP(&o.Placeholder, "placeholder", "", o.Placeholder, "The placeholder")
	fs.StringVarP(&o.TsUrl, "ts-url", "", o.TsUrl, "The ts url")
	fs.StringVarP(&o.TsToken, "ts-token", "", o.TsToken, "The ts token")
	fs.StringVarP(&o.CertFile, "cert-file", "", o.CertFile, "The Cert file")
	fs.StringVarP(&o.KeyFile, "key-file", "", o.KeyFile, "The Key file")
}

func (o *Options) Config(stopCh <-chan struct{}) (*config.Config, error) {

	c := &config.Config{}
	gatewayMgr := gateway.NewGatewayManager(stopCh)
	gatewayMgr.Init()
	c.GatewayMgr = gatewayMgr

	gatewayMeta, _ := gatewayMgr.GetGatewayMeta()
	store, _ := generic.NewStore(storage.StoreGroupToString[storage.StoreGroupDevice], storage.Devices, generic.DeviceTypeObjectMap)
	// redis
	client := redis.NewClient(&redis.Options{
		Addr:     o.RedisUrl,
		Password: o.Password,
		DB:       0,
	})
	// ts
	tsManager := ts.NewTsManager(o.TsUrl, o.TsToken)

	deviceMgr := device.NewManager(store, tsManager, client, gatewayMeta, o.Placeholder, stopCh)
	deviceMgr.Init()

	c.DeviceMgr = deviceMgr
	c.KeyFile = o.KeyFile
	c.CertFile = o.CertFile
	return c, nil
}
