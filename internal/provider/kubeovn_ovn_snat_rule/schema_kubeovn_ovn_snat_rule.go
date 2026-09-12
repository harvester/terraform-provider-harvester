package kubeovn_ovn_snat_rule

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNOvnSnatOvnEip: {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		constants.FieldKubeOVNOvnSnatVpcSubnet: {
			Type:     schema.TypeString,
			Required: true,
		},
		constants.FieldKubeOVNOvnSnatIPName: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNOvnSnatVpc: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNOvnSnatV4IpCidr: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNOvnSnatV6IpCidr: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNOvnSnatStatusReady: {
			Type:     schema.TypeBool,
			Computed: true,
		},
		constants.FieldKubeOVNOvnSnatStatusV4Eip: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNOvnSnatStatusV6Eip: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNOvnSnatStatusVpc: {
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
