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
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		constants.FieldKubeOVNVpcNatGwSubnet: {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		constants.FieldKubeOVNVpcNatGwLanIP: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		constants.FieldKubeOVNVpcNatGwExternalSubnets: {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		constants.FieldKubeOVNVpcNatGwSelector: {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		constants.FieldKubeOVNVpcNatGwQoSPolicy: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNVpcNatGwStatusQoS: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNVpcNatGwStatusExtSubs: {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		constants.FieldKubeOVNVpcNatGwStatusSelector: {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
