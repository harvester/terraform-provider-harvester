package constants

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
