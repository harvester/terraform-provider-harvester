package kubeovn_ovn_dnat_rule

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNOvnDnatOvnEip: {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		constants.FieldKubeOVNOvnDnatIPType: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNOvnDnatIPName: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNOvnDnatInternalPort: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNOvnDnatExternalPort: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNOvnDnatProtocol: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNOvnDnatVpc: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNOvnDnatV4IP: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNOvnDnatV6IP: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNOvnDnatStatusReady: {
			Type:     schema.TypeBool,
			Computed: true,
		},
		constants.FieldKubeOVNOvnDnatStatusV4Eip: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNOvnDnatStatusV6Eip: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNOvnDnatStatusVpc: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNOvnDnatStatusIPName: {
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
