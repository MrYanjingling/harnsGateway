package registry

import (
	"harns-platform/pkg/gateway/protocol/modbus"
	modbusallruntime "harns-platform/pkg/gateway/protocol/modbus/runtime"
	"harns-platform/pkg/gateway/protocol/opcua"
	opcuaruntime "harns-platform/pkg/gateway/protocol/opcua/runtime"
	"harns-platform/pkg/gateway/protocol/s7"
	s7runtime "harns-platform/pkg/gateway/protocol/s7/runtime"
	"harns-platform/pkg/gateway/runtime"
	v1 "harns-platform/pkg/gateway/v1"
)

var DeviceTypeMap = map[string]func() v1.DeviceType{
	"modbus": func() v1.DeviceType { return &v1.ModBusDevice{} },
	"opcUa":  func() v1.DeviceType { return &v1.OpcUaDevice{} },
	"s7":     func() v1.DeviceType { return &v1.S7Device{} },
}

var DeviceTypeObjectMap = map[string]runtime.Device{
	"modbus": &modbusallruntime.ModBusDevice{},
	"opcUa":  &opcuaruntime.OpcUaDevice{},
	"s7":     &s7runtime.S7Device{},
}

type NewBroker func(object runtime.Device) (runtime.Broker, chan *runtime.ParseVariableResult, error)

var DeviceTypeBrokerMap = map[string]NewBroker{
	"modbus": modbus.NewBroker,
	"opcUa":  opcua.NewBroker,
	"s7":     s7.NewBroker,
}
