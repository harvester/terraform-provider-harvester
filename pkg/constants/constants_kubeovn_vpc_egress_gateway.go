package constants

const (
	ResourceTypeKubeOVNVpcEgressGateway = "harvester_kubeovn_vpc_egress_gateway"

	FieldKubeOVNVpcEgressGatewayVpc            = "vpc"
	FieldKubeOVNVpcEgressGatewayReplicas       = "replicas"
	FieldKubeOVNVpcEgressGatewayPrefix         = "prefix"
	FieldKubeOVNVpcEgressGatewayImage          = "image"
	FieldKubeOVNVpcEgressGatewayInternalSubnet = "internal_subnet"
	FieldKubeOVNVpcEgressGatewayExternalSubnet = "external_subnet"
	FieldKubeOVNVpcEgressGatewayInternalIPs    = "internal_ips"
	FieldKubeOVNVpcEgressGatewayExternalIPs    = "external_ips"
	FieldKubeOVNVpcEgressGatewayTrafficPolicy  = "traffic_policy"

	// BFD config
	FieldKubeOVNVpcEgressGatewayBFD           = "bfd"
	FieldKubeOVNVpcEgressGatewayBFDEnabled    = "enabled"
	FieldKubeOVNVpcEgressGatewayBFDMinRX      = "min_rx"
	FieldKubeOVNVpcEgressGatewayBFDMinTX      = "min_tx"
	FieldKubeOVNVpcEgressGatewayBFDMultiplier = "multiplier"

	// Policies
	FieldKubeOVNVpcEgressGatewayPolicies       = "policy"
	FieldKubeOVNVpcEgressGatewayPolicySNAT     = "snat"
	FieldKubeOVNVpcEgressGatewayPolicyIPBlocks = "ip_blocks"
	FieldKubeOVNVpcEgressGatewayPolicySubnets  = "subnets"

	// Status
	FieldKubeOVNVpcEgressGatewayStatusReady       = "status_ready"
	FieldKubeOVNVpcEgressGatewayStatusPhase       = "status_phase"
	FieldKubeOVNVpcEgressGatewayStatusInternalIPs = "status_internal_ips"
	FieldKubeOVNVpcEgressGatewayStatusExternalIPs = "status_external_ips"
)
