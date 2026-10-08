package kubeovn_subnet

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNSubnetVpc: {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "ovn-cluster",
		},
		constants.FieldKubeOVNSubnetCIDRBlock: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.IsCIDR,
		},
		constants.FieldKubeOVNSubnetGateway: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.IsIPAddress,
		},
		constants.FieldKubeOVNSubnetExcludeIPs: {
			Type:        schema.TypeSet,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "IP addresses or ranges (\"10.0.0.10..10.0.0.20\") kept out of the subnet IPAM pool. kube-ovn always reserves the gateway, so it is added to the planned value when missing.",
		},
		constants.FieldKubeOVNSubnetProtocol: {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.StringInSlice([]string{kubeovnv1.ProtocolIPv4, kubeovnv1.ProtocolIPv6, kubeovnv1.ProtocolDual}, false),
			Description:  "IP family of the subnet. kube-ovn derives it from cidr_block and overwrites any other value, so it is computed when omitted.",
		},
		constants.FieldKubeOVNSubnetVlan: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNSubnetProvider: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNSubnetNamespaces: {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		constants.FieldKubeOVNSubnetEnableDHCP: {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		constants.FieldKubeOVNSubnetDHCPv4Options: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNSubnetPrivate: {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		constants.FieldKubeOVNSubnetAllowSubnets: {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		constants.FieldKubeOVNSubnetNatOutgoing: {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		constants.FieldKubeOVNSubnetGatewayType: {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "distributed",
			ValidateFunc: validation.StringInSlice([]string{"distributed", "centralized"}, false),
		},
		constants.FieldKubeOVNSubnetGatewayNode: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNSubnetEnableLb: {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		constants.FieldKubeOVNSubnetV4AvailableIPs: {
			Type:     schema.TypeFloat,
			Computed: true,
		},
		constants.FieldKubeOVNSubnetV4UsingIPs: {
			Type:     schema.TypeFloat,
			Computed: true,
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
