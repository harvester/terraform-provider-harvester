package kubeovn_vpc

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNVpcNamespaces: {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Namespaces allowed to use the subnets of the VPC.",
		},
		constants.FieldKubeOVNVpcStaticRoutes: {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Static routes of the VPC router.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldKubeOVNStaticRoutePolicy: {
						Type:         schema.TypeString,
						Optional:     true,
						Computed:     true,
						ValidateFunc: validation.StringInSlice([]string{string(kubeovnv1.PolicySrc), string(kubeovnv1.PolicyDst)}, false),
						Description:  "Match the destination (policyDst) or the source (policySrc) address. kube-ovn uses policyDst when it is omitted.",
					},
					constants.FieldKubeOVNStaticRouteCIDR: {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Destination or source CIDR of the route.",
					},
					constants.FieldKubeOVNStaticRouteNextHopIP: {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Next hop address.",
					},
					constants.FieldKubeOVNStaticRouteECMPMode: {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "ECMP mode of the route.",
					},
					constants.FieldKubeOVNStaticRouteTable: {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Route table of the route.",
					},
				},
			},
		},
		constants.FieldKubeOVNVpcPolicyRoutes: {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Policy routes of the VPC router.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldKubeOVNPolicyRoutePriority: {
						Type:        schema.TypeInt,
						Required:    true,
						Description: "Priority of the policy, the highest first.",
					},
					constants.FieldKubeOVNPolicyRouteMatch: {
						Type:        schema.TypeString,
						Required:    true,
						Description: "OVN match expression, for example \"ip4.src == 10.0.1.0/24\".",
					},
					constants.FieldKubeOVNPolicyRouteAction: {
						Type:     schema.TypeString,
						Required: true,
						ValidateFunc: validation.StringInSlice([]string{
							string(kubeovnv1.PolicyRouteActionAllow),
							string(kubeovnv1.PolicyRouteActionDrop),
							string(kubeovnv1.PolicyRouteActionReroute),
						}, false),
						Description: "allow, drop, or reroute to next_hop_ip.",
					},
					constants.FieldKubeOVNPolicyRouteNextHopIP: {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Next hop address of a reroute policy.",
					},
				},
			},
		},
		constants.FieldKubeOVNVpcEnableExternal: {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Connect the VPC router to the external network.",
		},
		constants.FieldKubeOVNVpcEnableBfd: {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Use BFD on the external connection of the VPC.",
		},
		constants.FieldKubeOVNVpcDefaultSubnet: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Default subnet of the VPC.",
		},
		constants.FieldKubeOVNVpcStandby: {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Whether kube-ovn has created the VPC router.",
		},
		constants.FieldKubeOVNVpcRouter: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Name of the OVN logical router of the VPC.",
		},
		constants.FieldKubeOVNVpcSubnets: {
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Subnets of the VPC.",
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
