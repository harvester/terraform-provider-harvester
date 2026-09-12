package kubeovn_subnet

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

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
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		constants.FieldKubeOVNSubnetProtocol: {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "IPv4",
			ValidateFunc: validation.StringInSlice([]string{"IPv4", "IPv6", "Dual"}, false),
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
