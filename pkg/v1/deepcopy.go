package v1

import "github.com/MrYanjingling/harns-platform/pkg/runtime"

func (d *DeviceMeta) DeepCopyObject() runtime.RunObject {
	out := *d
	return &out
}
