package kubeovn_iptables_fip_rule

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNIptablesFIPEIP: {
			Type:     schema.TypeString,
			Required: true,
		},
		constants.FieldKubeOVNIptablesFIPInternalIP: {
			Type:     schema.TypeString,
			Required: true,
		},
		constants.FieldKubeOVNIptablesFIPReady: {
			Type:     schema.TypeBool,
			Computed: true,
		},
		constants.FieldKubeOVNIptablesFIPStatusV4IP: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIptablesFIPStatusV6IP: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIptablesFIPStatusNat: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIptablesFIPStatusIP: {
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
