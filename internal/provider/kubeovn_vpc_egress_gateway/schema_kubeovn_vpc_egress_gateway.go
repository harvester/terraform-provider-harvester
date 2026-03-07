package kubeovn_vpc_egress_gateway

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNVpcEgressGatewayVpc: {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "VPC name. If not specified, the default VPC will be used",
		},
		constants.FieldKubeOVNVpcEgressGatewayReplicas: {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      1,
			ValidateFunc: validation.IntAtLeast(1),
			Description:  "Number of workload replicas",
		},
		constants.FieldKubeOVNVpcEgressGatewayPrefix: {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "Name prefix for the generated workload",
		},
		constants.FieldKubeOVNVpcEgressGatewayImage: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Image used by the workload. If not specified, the default kube-ovn image is used",
		},
		constants.FieldKubeOVNVpcEgressGatewayInternalSubnet: {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "Internal subnet for the workload. Defaults to the VPC's default subnet",
		},
		constants.FieldKubeOVNVpcEgressGatewayExternalSubnet: {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "External subnet for the workload",
		},
		constants.FieldKubeOVNVpcEgressGatewayInternalIPs: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "Internal IPs for the workload. Must be in the internal subnet. Count must not be less than replicas",
		},
		constants.FieldKubeOVNVpcEgressGatewayExternalIPs: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "External IPs for the workload. Must be in the external subnet. Count must not be less than replicas",
		},
		constants.FieldKubeOVNVpcEgressGatewayTrafficPolicy: {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "Cluster",
			ValidateFunc: validation.StringInSlice([]string{"Cluster", "Local"}, false),
			Description:  "Traffic routing policy: Cluster (default) or Local",
		},
		constants.FieldKubeOVNVpcEgressGatewayBFD: {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldKubeOVNVpcEgressGatewayBFDEnabled: {
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
						Description: "Enable BFD sessions with VPC BFD LRP",
					},
					constants.FieldKubeOVNVpcEgressGatewayBFDMinRX: {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "BFD minimum receive interval in milliseconds",
					},
					constants.FieldKubeOVNVpcEgressGatewayBFDMinTX: {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "BFD minimum transmit interval in milliseconds",
					},
					constants.FieldKubeOVNVpcEgressGatewayBFDMultiplier: {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "BFD detection multiplier",
					},
				},
			},
			Description: "BFD (Bidirectional Forwarding Detection) configuration",
		},
		constants.FieldKubeOVNVpcEgressGatewayPolicies: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldKubeOVNVpcEgressGatewayPolicySNAT: {
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
						Description: "Enable SNAT/MASQUERADE for egress traffic",
					},
					constants.FieldKubeOVNVpcEgressGatewayPolicyIPBlocks: {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Description: "CIDRs targeted by the egress traffic policy",
					},
					constants.FieldKubeOVNVpcEgressGatewayPolicySubnets: {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Description: "Subnets targeted by the egress traffic policy",
					},
				},
			},
			Description: "Egress traffic policies",
		},
		constants.FieldKubeOVNVpcEgressGatewayStatusReady: {
			Type:     schema.TypeBool,
			Computed: true,
		},
		constants.FieldKubeOVNVpcEgressGatewayStatusPhase: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNVpcEgressGatewayStatusInternalIPs: {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		constants.FieldKubeOVNVpcEgressGatewayStatusExternalIPs: {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
	util.NamespacedSchemaWrap(s, true)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
