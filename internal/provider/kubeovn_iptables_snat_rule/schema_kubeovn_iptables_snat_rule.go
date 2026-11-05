package kubeovn_iptables_snat_rule

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// Every spec field is ForceNew: kube-ovn refuses any spec change on a ready
// rule ("not support change").
func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNIptablesSnatEIP: {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Name of the iptables EIP whose address replaces the source address of the traffic.",
		},
		constants.FieldKubeOVNIptablesSnatInternalCIDR: {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "CIDR of the VPC addresses whose outgoing traffic is translated, usually the CIDR of the subnet.",
		},
		constants.FieldKubeOVNIptablesSnatStatusV4IP: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "IPv4 address of the EIP applied by kube-ovn.",
		},
		constants.FieldKubeOVNIptablesSnatStatusV6IP: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "IPv6 address of the EIP applied by kube-ovn.",
		},
		constants.FieldKubeOVNIptablesSnatStatusNat: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "NAT gateway the rule is applied in.",
		},
		constants.FieldKubeOVNIptablesSnatStatusCIDR: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Internal CIDR applied by kube-ovn.",
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
