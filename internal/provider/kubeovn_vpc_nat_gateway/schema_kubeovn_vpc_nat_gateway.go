package kubeovn_vpc_nat_gateway

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// vpc, subnet and lan_ip are ForceNew: the kube-ovn controller only rolls the
// gateway StatefulSet out again when the external subnets, the selector, the
// tolerations or the affinity change (isVpcNatGwChanged), so a new value of
// these fields would be stored but never applied to the running gateway.
func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNVpcNatGwVpc: {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Name of the VPC the gateway serves.",
		},
		constants.FieldKubeOVNVpcNatGwSubnet: {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Name of the VPC subnet the gateway pod is attached to.",
		},
		constants.FieldKubeOVNVpcNatGwLanIP: {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Address of the gateway in subnet. It must be inside the subnet CIDR and free: the VPC routes its external traffic to it.",
		},
		constants.FieldKubeOVNVpcNatGwExternalSubnets: {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Names of the external subnets of the gateway. kube-ovn attaches the gateway to the first one only and ignores the others; changing the first one rebuilds the gateway pod on that network. It must be backed by a network attachment definition: on Harvester, a kube-ovn subnet whose network_provider is <network name>.<network namespace>.ovn. kube-ovn uses the subnet ovn-vpc-external-network when it is omitted.",
		},
		constants.FieldKubeOVNVpcNatGwSelector: {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Node selector of the gateway pod, as \"key: value\" entries.",
		},
		constants.FieldKubeOVNVpcNatGwQoSPolicy: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Name of the kube-ovn QoS policy applied to the gateway.",
		},
		constants.FieldKubeOVNVpcNatGwStatusQoS: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "QoS policy applied by kube-ovn.",
		},
		constants.FieldKubeOVNVpcNatGwStatusExtSubs: {
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "External subnets applied by kube-ovn.",
		},
		constants.FieldKubeOVNVpcNatGwStatusSelector: {
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Node selector applied by kube-ovn.",
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
