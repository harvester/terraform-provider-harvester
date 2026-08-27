package virtualmachinetemplate

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldVirtualMachineTemplateDefaultVersionID: {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Default version ID in the format namespace/name. Automatically set by Harvester when the first version is created.",
		},
		constants.FieldVirtualMachineTemplateDefaultVersion: {
			Type:     schema.TypeInt,
			Computed: true,
		},
		constants.FieldVirtualMachineTemplateLatestVersion: {
			Type:     schema.TypeInt,
			Computed: true,
		},
	}
	util.NamespacedSchemaWrap(s, false)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
