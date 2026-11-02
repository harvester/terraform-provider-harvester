package constants

// KubeOVNDefaultVpc is the VPC kube-ovn creates for the default subnets
// (DefaultVpc in kube-ovn pkg/util, a package too heavy to import here).
const KubeOVNDefaultVpc = "ovn-cluster"

const (
	ResourceTypeKubeOVNSubnet = "harvester_kubeovn_subnet"

	FieldKubeOVNSubnetVpc            = "vpc"
	FieldKubeOVNSubnetCIDRBlock      = "cidr_block"
	FieldKubeOVNSubnetGateway        = "gateway"
	FieldKubeOVNSubnetExcludeIPs     = "exclude_ips"
	FieldKubeOVNSubnetProtocol       = "protocol"
	FieldKubeOVNSubnetVlan           = "vlan"
	FieldKubeOVNSubnetProvider       = "network_provider"
	FieldKubeOVNSubnetNamespaces     = "namespaces"
	FieldKubeOVNSubnetEnableDHCP     = "enable_dhcp"
	FieldKubeOVNSubnetDHCPv4Options  = "dhcp_v4_options"
	FieldKubeOVNSubnetPrivate        = "private"
	FieldKubeOVNSubnetAllowSubnets   = "allow_subnets"
	FieldKubeOVNSubnetNatOutgoing    = "nat_outgoing"
	FieldKubeOVNSubnetGatewayType    = "gateway_type"
	FieldKubeOVNSubnetGatewayNode    = "gateway_node"
	FieldKubeOVNSubnetEnableLb       = "enable_lb"
	FieldKubeOVNSubnetV4AvailableIPs = "v4_available_ips"
	FieldKubeOVNSubnetV4UsingIPs     = "v4_using_ips"
)
