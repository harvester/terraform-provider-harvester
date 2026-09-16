package kubeovn_iptables_dnat_rule

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNIptablesDnatEIP: {
			Type:     schema.TypeString,
			Required: true,
		},
		constants.FieldKubeOVNIptablesDnatExternalPort: {
			Type:     schema.TypeString,
			Required: true,
		},
		constants.FieldKubeOVNIptablesDnatProtocol: {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "tcp",
			ValidateFunc: validation.StringInSlice([]string{"tcp", "udp", "icmp"}, false),
		},
		constants.FieldKubeOVNIptablesDnatInternalIP: {
			Type:     schema.TypeString,
			Required: true,
		},
		constants.FieldKubeOVNIptablesDnatInternalPort: {
			Type:     schema.TypeString,
			Required: true,
		},
		constants.FieldKubeOVNIptablesDnatReady: {
			Type:     schema.TypeBool,
			Computed: true,
		},
		constants.FieldKubeOVNIptablesDnatStatusV4IP: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIptablesDnatStatusV6IP: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIptablesDnatStatusNat: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIptablesDnatStatusProto: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIptablesDnatStatusIntIP: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIptablesDnatStatusIntP: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIptablesDnatStatusExtP: {
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
