package addon

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldAddonEnabled: {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		constants.FieldAddonValuesContent: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		constants.FieldAddonRepo: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldAddonChart: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldAddonVersion: {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
	util.NamespacedSchemaWrap(s, true)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
