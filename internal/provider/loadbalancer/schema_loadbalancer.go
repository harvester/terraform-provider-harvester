package loadbalancer

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldLoadBalancerDescription: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldLoadBalancerWorkloadType: {
			Type:     schema.TypeString,
			Optional: true,
			ValidateFunc: validation.StringInSlice([]string{
				constants.LoadBalancerWorkloadTypeVM,
				constants.LoadBalancerWorkloadTypeCluster,
			}, false),
			Description: "Can be `vm` or `cluster`",
		},
		constants.FieldLoadBalancerIPAM: {
			Type:     schema.TypeString,
			Optional: true,
			ValidateFunc: validation.StringInSlice([]string{
				constants.LoadBalancerIPAMPool,
				constants.LoadBalancerIPAMDHCP,
			}, false),
			Description: "Where the load balancer gets its IP address from. Can be `dhcp` or `pool`.",
		},
		constants.FieldLoadBalancerIPPool: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Which IP pool to get the IP address from.",
		},
		constants.SubresourceTypeLoadBalancerListener: {
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Description: "",
			Elem: &schema.Resource{
				Schema: subresourceSchemaLoadBalancerListener(),
			},
		},
		constants.FieldLoadBalancerBackendServerSelector: {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "",
		},
		constants.SubresourceTypeLoadBalancerHealthCheck: {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "",
			Elem: &schema.Resource{
				Schema: subresourceSchemaLoadBalancerHealthCheck(),
			},
		},
	}
	util.NamespacedSchemaWrap(s, false)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}

func subresourceSchemaLoadBalancerListener() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldListenerName: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldListenerPort: {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "",
		},
		constants.FieldListenerProtocol: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "",
		},
		constants.FieldListenerBackendPort: {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "",
		},
	}
	return s
}

func subresourceSchemaLoadBalancerHealthCheck() map[string]*schema.Schema {
	s := map[string]*schema.Schema{}
	return s
}
