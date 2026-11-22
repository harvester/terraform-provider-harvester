package kubeovn_subnet

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNSubnetVpc: {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     constants.KubeOVNDefaultVpc,
			Description: "Name of the VPC the subnet belongs to.",
		},
		constants.FieldKubeOVNSubnetCIDRBlock: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validateDualStack(parseCIDR),
			Description:  "CIDR of the subnet. A dual-stack subnet takes an IPv4 and an IPv6 CIDR separated by a comma.",
		},
		constants.FieldKubeOVNSubnetGateway: {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validateDualStack(netip.ParseAddr),
			Description:  "Gateway address, one per IP family separated by a comma. kube-ovn uses the first address of cidr_block when it is omitted.",
		},
		constants.FieldKubeOVNSubnetExcludeIPs: {
			Type:        schema.TypeSet,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "IP addresses or ranges (\"10.0.0.10..10.0.0.20\") kept out of the subnet IPAM pool. kube-ovn always reserves the gateway and merges overlapping ranges, and the planned value does the same.",
		},
		constants.FieldKubeOVNSubnetProtocol: {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.StringInSlice([]string{kubeovnv1.ProtocolIPv4, kubeovnv1.ProtocolIPv6, kubeovnv1.ProtocolDual}, false),
			Description:  "IP family of the subnet. kube-ovn derives it from cidr_block and overwrites any other value, so it is computed when omitted.",
		},
		constants.FieldKubeOVNSubnetVlan: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Name of the kube-ovn VLAN of an underlay subnet.",
		},
		constants.FieldKubeOVNSubnetProvider: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Provider of the network attachment definition of the overlay network, in the form <nad name>.<nad namespace>.ovn. Harvester refuses a subnet without it.",
		},
		constants.FieldKubeOVNSubnetNamespaces: {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Namespaces bound to the subnet.",
		},
		constants.FieldKubeOVNSubnetEnableDHCP: {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Answer DHCP requests of the VMs on the subnet.",
		},
		constants.FieldKubeOVNSubnetDHCPv4Options: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "DHCPv4 options, for example \"lease_time=3600,router=10.0.0.1,server_id=169.254.0.254,server_mac=00:00:00:2E:2F:B8\".",
		},
		constants.FieldKubeOVNSubnetPrivate: {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Only accept traffic from the subnet itself and from allow_subnets.",
		},
		constants.FieldKubeOVNSubnetAllowSubnets: {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "CIDRs allowed to reach a private subnet.",
		},
		constants.FieldKubeOVNSubnetNatOutgoing: {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Masquerade the traffic that leaves the subnet.",
		},
		constants.FieldKubeOVNSubnetGatewayType: {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      kubeovnv1.GWDistributedType,
			ValidateFunc: validation.StringInSlice([]string{kubeovnv1.GWDistributedType, kubeovnv1.GWCentralizedType}, false),
			Description:  "Outgoing traffic leaves through every node (distributed) or through gateway_node (centralized).",
		},
		constants.FieldKubeOVNSubnetGatewayNode: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Nodes that carry the outgoing traffic of a centralized subnet, separated by commas.",
		},
		constants.FieldKubeOVNSubnetEnableLb: {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Apply the Kubernetes services load balancing on the subnet.",
		},
		constants.FieldKubeOVNSubnetV4AvailableIPs: {
			Type:        schema.TypeFloat,
			Computed:    true,
			Description: "Number of IPv4 addresses still free.",
		},
		constants.FieldKubeOVNSubnetV4UsingIPs: {
			Type:        schema.TypeFloat,
			Computed:    true,
			Description: "Number of IPv4 addresses in use.",
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}

func parseCIDR(s string) (netip.Addr, error) {
	prefix, err := netip.ParsePrefix(s)
	return prefix.Addr(), err
}

// validateDualStack accepts one value, or an IPv4 and an IPv6 value
// separated by a comma, as kube-ovn does for a dual-stack subnet.
func validateDualStack(parse func(string) (netip.Addr, error)) func(interface{}, string) ([]string, []error) {
	return func(i interface{}, key string) (warnings []string, errs []error) {
		values := strings.Split(i.(string), ",")
		if len(values) > 2 {
			return nil, []error{fmt.Errorf("%s: expected at most one IPv4 and one IPv6 value, got %q", key, i)}
		}
		families := map[bool]bool{}
		for _, value := range values {
			addr, err := parse(strings.TrimSpace(value))
			if err != nil {
				return nil, []error{fmt.Errorf("%s: %w", key, err)}
			}
			if families[addr.Is4()] {
				return nil, []error{fmt.Errorf("%s: expected at most one IPv4 and one IPv6 value, got %q", key, i)}
			}
			families[addr.Is4()] = true
		}
		return nil, nil
	}
}
