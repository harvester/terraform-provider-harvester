package kubeovn_iptables_snat_rule

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNIptablesSnatEIP: {
			Type:     schema.TypeString,
			Required: true,
		},
		constants.FieldKubeOVNIptablesSnatInternalCIDR: {
			Type:     schema.TypeString,
			Required: true,
		},
		constants.FieldKubeOVNIptablesSnatReady: {
			Type:     schema.TypeBool,
			Computed: true,
		},
		constants.FieldKubeOVNIptablesSnatStatusV4IP: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIptablesSnatStatusV6IP: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIptablesSnatStatusNat: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIptablesSnatStatusCIDR: {
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
