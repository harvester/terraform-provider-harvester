package kubeovn_iptables_dnat_rule

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// Every spec field is ForceNew: kube-ovn refuses any spec change on a ready
// rule ("not support change").
func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNIptablesDnatEIP: {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Name of the iptables EIP whose address receives the traffic.",
		},
		constants.FieldKubeOVNIptablesDnatExternalPort: {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Port of the EIP the traffic arrives on.",
		},
		constants.FieldKubeOVNIptablesDnatProtocol: {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Default:     constants.KubeOVNIptablesDnatProtocolTCP,
			Description: "Protocol of the forwarded traffic, tcp or udp.",
		},
		constants.FieldKubeOVNIptablesDnatInternalIP: {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Address in the VPC the traffic is forwarded to.",
		},
		constants.FieldKubeOVNIptablesDnatInternalPort: {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Port of internal_ip the traffic is forwarded to.",
		},
		constants.FieldKubeOVNIptablesDnatStatusV4IP: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "IPv4 address of the EIP applied by kube-ovn.",
		},
		constants.FieldKubeOVNIptablesDnatStatusV6IP: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "IPv6 address of the EIP applied by kube-ovn.",
		},
		constants.FieldKubeOVNIptablesDnatStatusNat: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "NAT gateway the rule is applied in.",
		},
		constants.FieldKubeOVNIptablesDnatStatusProto: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Protocol applied by kube-ovn.",
		},
		constants.FieldKubeOVNIptablesDnatStatusIntIP: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Internal address applied by kube-ovn.",
		},
		constants.FieldKubeOVNIptablesDnatStatusIntP: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Internal port applied by kube-ovn.",
		},
		constants.FieldKubeOVNIptablesDnatStatusExtP: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "External port applied by kube-ovn.",
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
