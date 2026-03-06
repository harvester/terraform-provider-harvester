package kubeovn_ovn_fip

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNOvnFipOvnEip: {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		constants.FieldKubeOVNOvnFipIPType: {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{"vip", "ip"}, false),
		},
		constants.FieldKubeOVNOvnFipIPName: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNOvnFipVpc: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNOvnFipV4IP: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNOvnFipV6IP: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNOvnFipStatusReady: {
			Type:     schema.TypeBool,
			Computed: true,
		},
		constants.FieldKubeOVNOvnFipStatusV4Eip: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNOvnFipStatusV6Eip: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNOvnFipStatusV4IP: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNOvnFipStatusV6IP: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNOvnFipStatusVpc: {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
