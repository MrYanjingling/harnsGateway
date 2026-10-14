package v1

import "harns-platform/pkg/gateway/runtime"

func (d *DeviceMeta) DeepCopyObject() runtime.RunObject {
	out := *d
	return &out
}
