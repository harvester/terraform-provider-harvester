package virtualmachine

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldVirtualMachineMachineType: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		constants.FieldVirtualMachineHostname: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		constants.FieldVirtualMachineStart: {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		constants.FieldVirtualMachineCPU: {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  1,
		},
		constants.FieldVirtualMachineMemory: {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "1Gi",
		},
		constants.FieldVirtualMachineSSHKeys: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		constants.FieldVirtualMachineCloudInit: {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: resourceCloudInitSchema(),
			},
		},
		constants.FieldVirtualMachineDisk: {
			Type:     schema.TypeList,
			Required: true,
			MinItems: 1,
			Elem: &schema.Resource{
				Schema: resourceDiskSchema(),
			},
		},
		constants.FieldVirtualMachineNetworkInterface: {
			Type:     schema.TypeList,
			Required: true,
			MinItems: 1,
			Elem: &schema.Resource{
				Schema: resourceNetworkInterfaceSchema(),
			},
		},
		constants.FieldVirtualMachineInstanceNodeName: {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
	util.NamespacedSchemaWrap(s, false)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
