package sriovdevice

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldSRIOVNetworkDeviceNumVFs: {
			Type:        schema.TypeInt,
			Computed:    true,
			Optional:    true,
			Description: "Number of Virtual Functions (VFs)",
		},
		constants.FieldSRIOVNetworkDeviceEnabled: {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "SRIOV enabled",
		},
		constants.FieldSRIOVNetworkDeviceVFAddresses: {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "PCI bus addresses of the virtual functions",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		constants.FieldSRIOVNetworkDeviceVFDeviceNames: {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "Names of the PCI devices of the virtual functions",
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
