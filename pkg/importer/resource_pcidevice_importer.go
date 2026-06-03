package importer

import (
	devicesv1 "github.com/harvester/pcidevices/pkg/apis/devices.harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourcePCIDeviceStateGetter(obj *devicesv1.PCIDevice) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:                 obj.Name,
		constants.FieldPCIDeviceDeviceDescription: obj.Status.Description,
		constants.FieldPCIDeviceAddress:           obj.Status.Address,
		constants.FieldPCIDeviceNodeName:          obj.Status.NodeName,
		constants.FieldPCIDeviceVendorID:          obj.Status.VendorID,
		constants.FieldPCIDeviceDeviceID:          obj.Status.DeviceID,
		constants.FieldPCIDeviceClassID:           obj.Status.ClassID,
		constants.FieldPCIDeviceIOMMUGroup:        obj.Status.IOMMUGroup,
		constants.FieldPCIDeviceKernelDriver:      obj.Status.KernelDriverInUse,
		constants.FieldPCIDeviceResourceName:      obj.Status.ResourceName,
	}
	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypePCIDevice,
		States:       states,
	}, nil
}
