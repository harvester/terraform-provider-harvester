package resourcequota

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldResourceQuotaNamespaceTotalSnapshotSizeQuota: {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Total snapshot size quota for the namespace in bytes",
		},
		constants.FieldResourceQuotaVMTotalSnapshotSizeQuota: {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "Per-VM snapshot size quotas in bytes",
			Elem: &schema.Schema{
				Type: schema.TypeInt,
			},
		},
		constants.FieldResourceQuotaNamespaceTotalSnapshotSizeUsage: {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Current total snapshot size usage for the namespace in bytes",
		},
		constants.FieldResourceQuotaVMTotalSnapshotSizeUsage: {
			Type:        schema.TypeMap,
			Computed:    true,
			Description: "Current per-VM snapshot size usage in bytes",
			Elem: &schema.Schema{
				Type: schema.TypeInt,
			},
		},
	}
	util.NamespacedSchemaWrap(s, false)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
