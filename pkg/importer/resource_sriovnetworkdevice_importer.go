package importer

import (
	devicesv1 "github.com/harvester/pcidevices/pkg/apis/devices.harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceSRIOVNetworkDeviceStateGetter(obj *devicesv1.SRIOVNetworkDevice) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:                      obj.Name,
		constants.FieldSRIOVNetworkDeviceNumVFs:        obj.Spec.NumVFs,
		constants.FieldSRIOVNetworkDeviceEnabled:       obj.Status.Status == devicesv1.DeviceEnabled,
		constants.FieldSRIOVNetworkDeviceVFAddresses:   obj.Status.VFAddresses,
		constants.FieldSRIOVNetworkDeviceVFDeviceNames: obj.Status.VFPCIDevices,
	}
	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeSRIOVNetworkDevice,
		States:       states,
	}, nil
}
