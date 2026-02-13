package pcidevice

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

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
			Description: "The node where the PCI devices are available. The VM must run on this node.",
		},
		constants.FieldPCIDevicePCIAddresses: {
			Type:     schema.TypeList,
			Required: true,
			MinItems: 1,
			Description: "List of PCI addresses to passthrough (format: '0000:XX:YY.Z')." +
				" Devices must be enabled for passthrough in Harvester.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^[0-9a-fA-F]{4}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}\.[0-9a-fA-F]$`),
					"PCI address must be in format '0000:XX:YY.Z' (e.g., '0000:01:00.0')",
				),
			},
		},
	}
	util.NamespacedSchemaWrap(s, false)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
