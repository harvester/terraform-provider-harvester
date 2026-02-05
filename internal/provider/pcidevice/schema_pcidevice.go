// Package pcidevice provides the Terraform schema definitions for the harvester_pci_device resource.
package pcidevice

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// Schema returns the Terraform schema for the harvester_pci_device resource.
// This resource manages PCI device passthrough to VMs using Harvester's PCIDeviceClaim CRD.
// According to Harvester documentation, the VM must be scheduled on a specific node where
// the PCI devices are available to avoid scheduling issues.
func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldPCIDeviceVMName: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the virtual machine to attach PCI devices to. Format: 'namespace/name' or 'name' (if in default namespace).",
		},
		constants.FieldPCIDeviceNodeName: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the node where the VM must be deployed. This is mandatory to ensure the VM runs on the correct node where the PCI devices are available. This prevents scheduling issues when multiple nodes have the same PCI device type.",
		},
		constants.FieldPCIDevicePCIAddresses: {
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Description: "List of PCI addresses to attach to the VM. Each address should be in the format '0000:XX:YY.Z' (e.g., '0000:01:00.0'). The PCI devices must be enabled for passthrough in Harvester before they can be attached.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^[0-9a-fA-F]{4}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}\.[0-9a-fA-F]$`),
					"PCI address must be in format '0000:XX:YY.Z' (e.g., '0000:01:00.0')",
				),
			},
		},
		constants.FieldPCIDeviceLabels: {
			Type:        schema.TypeMap,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Labels to apply to the PCIDeviceClaim resource.",
		},
	}
	util.NamespacedSchemaWrap(s, false)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}

