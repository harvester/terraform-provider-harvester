package vgpudevice

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		// --- read/write properties ---
		constants.FieldVGPUDeviceEnabled: {
			Type:        schema.TypeBool,
			Description: "Enable vGPU for passthrough",
			Optional:    true,
			Computed:    true,
		},
		constants.FieldVGPUDeviceType: {
			Type:        schema.TypeString,
			Description: "Configured vGPU type for this vGPU device",
			Optional:    true,
		},
		constants.FieldVGPUDeviceParentGPUDeviceAddress: {
			Type:        schema.TypeString,
			Description: "PCI bus address of parent GPU device",
			Optional:    true,
		},
		constants.FieldVGPUDeviceNodeName: {
			Type:        schema.TypeString,
			Description: "Name of node where parent GPU device is located",
			Optional:    true,
		},
		// --- read-only status properties ---
		constants.FieldVGPUDeviceStatus: {
			Type:        schema.TypeString,
			Description: "Current configuration status of vGPU device",
			Computed:    true,
		},
		constants.FieldVGPUDeviceUUID: {
			Type:        schema.TypeString,
			Description: "UUID of vGPU device",
			Computed:    true,
		},
		constants.FieldVGPUDeviceConfiguredVGPUType: {
			Type:        schema.TypeString,
			Description: "Current configured vGPU device type",
			Computed:    true,
		},
		constants.FieldVGPUDeviceAvailableTypes: {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "Available vGPU types for this vGPU device",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
