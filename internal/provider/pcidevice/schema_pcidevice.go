package pcidevice

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldPCIDevicePassthroughEnabled: {
			Type:        schema.TypeBool,
			Computed:    true,
			Optional:    true,
			Description: "Enable/Disable PCI passthrough",
		},
		constants.FieldPCIDeviceAddress: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "PCI address (e.g., '0000:01:00.0').",
		},
		constants.FieldPCIDeviceNodeName: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Node where the PCI device is located.",
		},
		constants.FieldPCIDeviceVendorID: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "PCI vendor ID (e.g., '8086' for Intel).",
		},
		constants.FieldPCIDeviceDeviceID: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "PCI device ID.",
		},
		constants.FieldPCIDeviceClassID: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "PCI class ID (e.g., '0300' for VGA).",
		},
		constants.FieldPCIDeviceDeviceDescription: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Human-readable device description.",
		},
		constants.FieldPCIDeviceIOMMUGroup: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "IOMMU group of the device.",
		},
		constants.FieldPCIDeviceKernelDriver: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Kernel driver currently in use.",
		},
		constants.FieldPCIDeviceResourceName: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Kubernetes device plugin resource name.",
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
