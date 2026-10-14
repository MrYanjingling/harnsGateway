package options

import (
	"fmt"
	"github.com/MrYanjingling/harns-platform/cmd/gateway/config"
	"github.com/MrYanjingling/harns-platform/pkg/device"
	"github.com/MrYanjingling/harns-platform/pkg/gateway"
	"github.com/MrYanjingling/harns-platform/pkg/generic"
	baseoptions "github.com/MrYanjingling/harns-platform/pkg/generic/options"
	"github.com/MrYanjingling/harns-platform/pkg/logging"
	"github.com/MrYanjingling/harns-platform/pkg/storage"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/pflag"
	"time"
)

type Options struct {
	Port           string               `json:"port"`
	Wait           baseoptions.Duration `json:"graceful-timeout"`
	MqttBrokerUrls []string             `json:"mqtt-broker-urls"`
	MqttUsername   string               `json:"mqtt-username"`
	MqttPassword   string               `json:"mqtt-password"`
	CertFile       string               `json:"cert-file"`
	KeyFile        string               `json:"key-file"`
	baseoptions.BaseOptions
	// logs.BaseOptions
}

const (
	_defaultPort         = "32200"
	_defaultWait         = baseoptions.Duration(15 * time.Second)
	_defaultMqttUsername = ""
	_defaultMqttPassword = ""
	_defaultConfigFile   = "/etc/gateway/config.yaml"
)

var (
	_defaultMqttBrokerUrls = []string{"tcp://127.0.0.1:1883"}
)

func NewDefaultOptions() *Options {
	base := baseoptions.NewDefaultBaseOptions()
	base.ConfigFile = _defaultConfigFile
	return &Options{
		Port:           _defaultPort,
		Wait:           _defaultWait,
		MqttBrokerUrls: _defaultMqttBrokerUrls,
		MqttUsername:   _defaultMqttUsername,
		MqttPassword:   _defaultMqttPassword,
		BaseOptions:    base,
		CertFile:       "",
		KeyFile:        "",
		// BaseOptions: logs.NewOptions(),
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	// refer to node port assignment https://rancher.com/docs/rancher/v2.x/en/installation/requirements/ports/#commonly-used-ports
	fs.StringVarP(&o.Port, "port", "P", o.Port, "Port exposed")
	baseoptions.DurationVar(fs, &o.Wait, "graceful-timeout", "The duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	fs.StringSliceVarP(&o.MqttBrokerUrls, "mqtt-broker-urls", "", o.MqttBrokerUrls, "The MQTT device urls. The format should be scheme://host:port Where \"scheme\" is one of \"tcp\", \"ssl\", or \"ws\"")
	fs.StringVarP(&o.MqttUsername, "mqtt-username", "u", o.MqttUsername, "The MQTT username")
	fs.StringVarP(&o.MqttPassword, "mqtt-password", "p", o.MqttPassword, "The MQTT password")
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
	// mqtt
	mqttOption := mqtt.NewClientOptions()
	for _, s := range o.MqttBrokerUrls {
		mqttOption = mqttOption.AddBroker(s)
	}
	mqttOption.SetUsername(o.MqttUsername)
	mqttOption.SetPassword(o.MqttPassword)
	mqttOption.SetOrderMatters(false)
	mqttOption.SetClientID(fmt.Sprintf("gateway-id-%s", gatewayMeta.ID))
	mqttClient := mqtt.NewClient(mqttOption)
	logging.V(1).InfoS("Connected to MQTT", "servers", o.MqttBrokerUrls)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		logging.ErrorS(token.Error(), "Failed to connect MQTT", "servers", o.MqttBrokerUrls)
		return nil, token.Error()
	}
	deviceMgr := device.NewManager(store, mqttClient, gatewayMeta, stopCh)
	deviceMgr.Init()

	c.DeviceMgr = deviceMgr
	c.KeyFile = o.KeyFile
	c.CertFile = o.CertFile
	return c, nil
}
