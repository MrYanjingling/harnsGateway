package generic

import (
	"github.com/MrYanjingling/harns-platform/pkg/protocol/modbus"
	modbusallruntime "github.com/MrYanjingling/harns-platform/pkg/protocol/modbus/runtime"
	"github.com/MrYanjingling/harns-platform/pkg/protocol/opcua"
	opcuaruntime "github.com/MrYanjingling/harns-platform/pkg/protocol/opcua/runtime"
	"github.com/MrYanjingling/harns-platform/pkg/protocol/s7"
	s7runtime "github.com/MrYanjingling/harns-platform/pkg/protocol/s7/runtime"
	"github.com/MrYanjingling/harns-platform/pkg/runtime"
	v1 "github.com/MrYanjingling/harns-platform/pkg/v1"
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
