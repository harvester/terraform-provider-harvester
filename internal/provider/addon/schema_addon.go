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
			Description: "Whether the addon is deployed. This is declarative: adopting an " +
				"already enabled addon without setting enabled = true disables it.",
		},
		constants.FieldAddonValuesContent: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			Description: "Helm values (YAML) applied to the addon chart. Once set it cannot " +
				"be reset to empty from Terraform, overwrite it with new values instead.",
		},
		constants.FieldAddonRepo: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			Description: "Helm repository URL of the addon chart. Read from the cluster for " +
				"built-in addons; required together with chart and version to create a custom addon.",
		},
		constants.FieldAddonChart: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			Description: "Helm chart name of the addon. Immutable on an existing addon " +
				"(enforced by the Harvester webhook).",
		},
		constants.FieldAddonVersion: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			Description: "Helm chart version of the addon. Read from the cluster for " +
				"built-in addons; required together with repo and chart to create a custom addon.",
		},
	}
	util.NamespacedSchemaWrap(s, true)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
