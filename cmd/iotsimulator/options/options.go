package options

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/pflag"
	"harnsgateway/cmd/iotsimulator/config"
	baseoptions "harnsgateway/pkg/generic/options"
	"harnsgateway/pkg/iotsimulator"
)

type Options struct {
	MqttBrokerUrls []string `json:"mqtt-broker-urls"`
	MqttUsername   string   `json:"mqtt-username"`
	MqttPassword   string   `json:"mqtt-password"`
	DeviceNumber   int      `json:"device-number"`
	KeyNumber      int      `json:"key-number"`
	baseoptions.BaseOptions
	// logs.BaseOptions
}

const (
	_defaultMqttUsername = "ddc@Dcom"
	_defaultMqttPassword = "ddc^Dcom_pwd"
	_defaultKeyNumber    = 20
	_defaultDeviceNumber = 1000
)

var (
	_defaultMqttBrokerUrls = []string{"tcp://10.56.223.141:8793"}
)

func NewDefaultOptions() *Options {
	return &Options{
		MqttBrokerUrls: _defaultMqttBrokerUrls,
		MqttUsername:   _defaultMqttUsername,
		MqttPassword:   _defaultMqttPassword,
		DeviceNumber:   _defaultDeviceNumber,
		KeyNumber:      _defaultKeyNumber,
		BaseOptions:    baseoptions.NewDefaultBaseOptions(),
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	// refer to node port assignment https://rancher.com/docs/rancher/v2.x/en/installation/requirements/ports/#commonly-used-ports
	fs.StringSliceVarP(&o.MqttBrokerUrls, "mqtt-broker-urls", "", o.MqttBrokerUrls, "The MQTT device urls. The format should be scheme://host:port Where \"scheme\" is one of \"tcp\", \"ssl\", or \"ws\"")
	fs.StringVarP(&o.MqttUsername, "mqtt-username", "u", o.MqttUsername, "The MQTT username")
	fs.StringVarP(&o.MqttPassword, "mqtt-password", "p", o.MqttPassword, "The MQTT password")
	fs.IntVarP(&o.DeviceNumber, "device-number", "d", o.DeviceNumber, "The Device number")
	fs.IntVarP(&o.KeyNumber, "key-number", "k", o.KeyNumber, "The Device number")
}

func (o *Options) Config(stopCh <-chan struct{}) (*config.Config, error) {
	c := &config.Config{}
	ismm := make(map[string]*mqtt.ClientOptions, o.DeviceNumber) // clientId -> clientOption

	for i := 1; i < o.DeviceNumber+1; i++ {
		mqttOption := mqtt.NewClientOptions()
		for _, s := range o.MqttBrokerUrls {
			mqttOption = mqttOption.AddBroker(s)
		}
		mqttOption.SetUsername(o.MqttUsername)
		mqttOption.SetPassword(o.MqttPassword)
		mqttOption.SetOrderMatters(false)
		clientId := fmt.Sprintf("iot-simulator-%d", i)
		mqttOption.SetClientID(clientId)
		ismm[clientId] = mqttOption
	}

	iotSimulatorMgr := iotsimulator.NewIotSimulatorManager(stopCh, ismm, o.KeyNumber, o.DeviceNumber)
	iotSimulatorMgr.Init()
	c.IotSimulatorMgr = iotSimulatorMgr
	return c, nil
}
