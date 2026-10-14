package device

import (
	"github.com/MrYanjingling/harns-platform/pkg/runtime"
	v1 "github.com/MrYanjingling/harns-platform/pkg/v1"
)

type DeviceManager interface {
	CreateDevice(deviceType v1.DeviceType) (runtime.Device, error)
	DeleteDevice(device runtime.Device) (runtime.Device, error)
	UpdateValidation(deviceType v1.DeviceType, device runtime.Device) error
	UpdateDevice(id string, deviceType v1.DeviceType, device runtime.Device) (runtime.Device, error)
}
