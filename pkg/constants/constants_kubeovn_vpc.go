package constants

const (
	ResourceTypeKubeOVNVpc = "harvester_kubeovn_vpc"

	FieldKubeOVNVpcNamespaces       = "namespaces"
	FieldKubeOVNVpcStaticRoutes     = "static_routes"
	FieldKubeOVNVpcPolicyRoutes     = "policy_routes"
	FieldKubeOVNVpcEnableExternal   = "enable_external"
	FieldKubeOVNVpcEnableBfd        = "enable_bfd"
	FieldKubeOVNVpcDefaultSubnet    = "default_subnet"
	FieldKubeOVNVpcStandby          = "standby"
	FieldKubeOVNVpcRouter           = "router"
	FieldKubeOVNVpcSubnets          = "subnets"
	FieldKubeOVNVpcEnableExternalSt = "status_enable_external"

	FieldKubeOVNStaticRoutePolicy    = "policy"
	FieldKubeOVNStaticRouteCIDR      = "cidr"
	FieldKubeOVNStaticRouteNextHopIP = "next_hop_ip"
	FieldKubeOVNStaticRouteECMPMode  = "ecmp_mode"
	FieldKubeOVNStaticRouteTable     = "route_table"

	FieldKubeOVNPolicyRoutePriority  = "priority"
	FieldKubeOVNPolicyRouteMatch     = "match"
	FieldKubeOVNPolicyRouteAction    = "action"
	FieldKubeOVNPolicyRouteNextHopIP = "next_hop_ip"
)
