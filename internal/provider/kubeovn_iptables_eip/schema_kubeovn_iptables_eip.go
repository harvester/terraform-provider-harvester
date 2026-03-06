package kubeovn_iptables_eip

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNIptablesEIPV4IP: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		constants.FieldKubeOVNIptablesEIPV6IP: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		constants.FieldKubeOVNIptablesEIPMacAddress: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		constants.FieldKubeOVNIptablesEIPNatGwDp: {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		constants.FieldKubeOVNIptablesEIPQoSPolicy: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNIptablesEIPExternalSubnet: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNIptablesEIPReady: {
			Type:     schema.TypeBool,
			Computed: true,
		},
		constants.FieldKubeOVNIptablesEIPStatusIP: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIptablesEIPStatusNat: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIptablesEIPStatusQoS: {
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
