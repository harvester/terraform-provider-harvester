package importer

import (
	devicesv1 "github.com/harvester/pcidevices/pkg/apis/devices.harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceSRIOVGPUDeviceStateGetter(obj *devicesv1.SRIOVGPUDevice) (*StateGetter, error) {
	states := map[string]any{
		constants.FieldCommonName:                  obj.Name,
		constants.FieldSRIOVGPUDeviceEnabled:       obj.Spec.Enabled,
		constants.FieldSRIOVGPUDeviceVFAddresses:   obj.Status.VFAddresses,
		constants.FieldSRIOVGPUDeviceVFDeviceNames: obj.Status.VGPUDevices,
	}
	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeSRIOVGPUDevice,
		States:       states,
	}, nil
}
