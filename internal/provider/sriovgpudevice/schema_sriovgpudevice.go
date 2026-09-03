package sriovgpudevice

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldSRIOVGPUDeviceEnabled: {
			Type:        schema.TypeBool,
			Description: "SRIOV enabled",
			Optional:    true,
			Computed:    true,
		},
		constants.FieldSRIOVGPUDeviceVFAddresses: {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "PCI bus addresses of the virtual functions",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		constants.FieldSRIOVGPUDeviceVFDeviceNames: {
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
