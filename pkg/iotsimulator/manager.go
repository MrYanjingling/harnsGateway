package iotsimulator

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"harnsgateway/pkg/runtime"
	"k8s.io/klog/v2"
	"math/rand"
	"time"
)

type Manager struct {
	Ismm         map[string]*mqtt.ClientOptions
	Stop         <-chan struct{}
	KeyNumber    int
	DeviceNumber int
	TopicPrefix  string
}

func NewIotSimulatorManager(stop <-chan struct{}, ismm map[string]*mqtt.ClientOptions, keyNumber int, deviceNumber int) *Manager {

	return &Manager{
		Stop:         stop,
		Ismm:         ismm,
		KeyNumber:    keyNumber,
		DeviceNumber: deviceNumber,
		TopicPrefix:  "data/v1/",
	}
}

func (m *Manager) Init() {
	for ci, mq := range m.Ismm {
		c := mqtt.NewClient(mq)
		if token := c.Connect(); token.Wait() && token.Error() != nil {
			klog.ErrorS(token.Error(), "Failed to connect MQTT", "err", token.Error())
		}
		m.Poll(ci, c)
	}
}

func (m *Manager) Poll(clientId string, client mqtt.Client) {
	ticker := time.NewTicker(1 * time.Second)
	// defer ticker.Stop()
	// done := make(chan bool)
	go func() {
		for {
			select {
			case <-ticker.C:
				m.PublishDataToMqtt(clientId, client)
			case <-m.Stop:
				// done <- true
				klog.V(1).InfoS("Exit poll data")
				return
			}
		}
	}()
	// <-done

}

func (m *Manager) PublishDataToMqtt(clientId string, client mqtt.Client) {

	topic := m.TopicPrefix + clientId
	publishData := runtime.PublishData{Payload: runtime.Payload{Data: []runtime.TimeSeriesData{{
		Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		Values:    m.GenerateData(),
	}}}}

	marshal, _ := json.Marshal(publishData)
	token := client.Publish(topic, 1, true, marshal)
	if token.WaitTimeout(3*time.Second) && token.Error() == nil {
		klog.V(5).InfoS("Succeed to publish MQTT", "topic", topic, "data", publishData)
	} else {
		klog.V(1).InfoS("Failed to publish MQTT", "topic", topic, "err", token.Error())
	}

}

func (m *Manager) GenerateData() []runtime.PointData {
	pds := make([]runtime.PointData, 0, m.KeyNumber)
	for i := 1; i < m.KeyNumber+1; i++ {
		pd := runtime.PointData{
			DataPointId: fmt.Sprintf("key-%d", i),
			Value:       rand.Int31n(998),
		}
		pds = append(pds, pd)
	}
	return pds
}
