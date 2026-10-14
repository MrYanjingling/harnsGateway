package model

import (
	modbus "harns-platform/pkg/gateway/protocol/modbus/runtime"
	"harns-platform/pkg/gateway/runtime/constant"
)

var _ ModbusModeler = (*ModbusTcp)(nil)
var _ ModbusModeler = (*ModbusRtu)(nil)
var _ ModbusModeler = (*ModbusRtuOverTcp)(nil)

var ModbusModelers = map[string]ModbusModeler{
	"modbusTcp":        &ModbusTcp{},
	"modbusRtu":        &ModbusRtu{},
	"modbusRtuOverTcp": &ModbusRtuOverTcp{},
}

type ModbusModeler interface {
	GenerateReadMessage(slave uint, functionCode uint8, startAddress uint, maxDataSize uint, variables []*modbus.VariableParse, memoryLayout constant.MemoryLayout) *modbus.ModBusDataFrame
	NewClients(address *modbus.Address, dataFrameCount int) (*modbus.Clients, error)
	// ExecuteAction(messenger modbus.Messenger, variables []*modbus.Variable) error
}
