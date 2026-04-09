package namespace

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldNamespaceDeleteOnDestroy: {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "If true, the namespace will be deleted on terraform destroy. Default is false to prevent accidental deletion of namespaces containing resources.",
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
