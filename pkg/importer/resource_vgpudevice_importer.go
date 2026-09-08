package importer

import (
	devicesv1 "github.com/harvester/pcidevices/pkg/apis/devices.harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceVGPUDeviceStateGetter(obj *devicesv1.VGPUDevice) (*StateGetter, error) {
	states := map[string]any{
		constants.FieldCommonName:                       obj.Name,
		constants.FieldVGPUDeviceEnabled:                obj.Spec.Enabled,
		constants.FieldVGPUDeviceType:                   obj.Spec.VGPUTypeName,
		constants.FieldVGPUDeviceParentGPUDeviceAddress: obj.Spec.ParentGPUDeviceAddress,
		constants.FieldVGPUDeviceNodeName:               obj.Spec.NodeName,
		constants.FieldVGPUDeviceStatus:                 obj.Status.VGPUStatus,
		constants.FieldVGPUDeviceUUID:                   obj.Status.UUID,
		constants.FieldVGPUDeviceConfiguredVGPUType:     obj.Status.ConfiguredVGPUTypeName,
		constants.FieldVGPUDeviceAvailableTypes:         obj.Status.AvailableTypes,
	}

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeSRIOVGPUDevice,
		States:       states,
	}, nil
}
