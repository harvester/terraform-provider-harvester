package kubeovn_iptables_eip

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// Every spec field is ForceNew: kube-ovn refuses any spec change on a ready
// EIP ("not support change").
func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNIptablesEIPV4IP: {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Computed:    true,
			Description: "IPv4 address of the EIP, inside the CIDR of external_subnet. kube-ovn allocates a free one when it is omitted.",
		},
		constants.FieldKubeOVNIptablesEIPV6IP: {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Computed:    true,
			Description: "IPv6 address of the EIP, in lower case and inside the CIDR of external_subnet.",
		},
		constants.FieldKubeOVNIptablesEIPMacAddress: {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Computed:    true,
			Description: "MAC address of the EIP. kube-ovn generates one when it is omitted.",
		},
		constants.FieldKubeOVNIptablesEIPNatGwDp: {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Name of the VPC NAT gateway the EIP is attached to.",
		},
		constants.FieldKubeOVNIptablesEIPQoSPolicy: {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "Name of the kube-ovn QoS policy applied to the EIP.",
		},
		constants.FieldKubeOVNIptablesEIPExternalSubnet: {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "Name of the external subnet the address is taken from, one of the external_subnets of the NAT gateway. kube-ovn uses the subnet ovn-vpc-external-network when it is omitted, which Harvester does not create.",
		},
		constants.FieldKubeOVNIptablesEIPStatusIP: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Address of the EIP applied by kube-ovn.",
		},
		constants.FieldKubeOVNIptablesEIPStatusNat: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Comma-separated kinds of NAT rules using the EIP (snat, dnat, fip).",
		},
		constants.FieldKubeOVNIptablesEIPStatusQoS: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "QoS policy applied by kube-ovn.",
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
