package constants

const (
	ResourceTypeKubeOVNIptablesDnatRule = "harvester_kubeovn_iptables_dnat_rule"

	FieldKubeOVNIptablesDnatEIP          = "eip"
	FieldKubeOVNIptablesDnatExternalPort = "external_port"
	FieldKubeOVNIptablesDnatProtocol     = "protocol"
	FieldKubeOVNIptablesDnatInternalIP   = "internal_ip"
	FieldKubeOVNIptablesDnatInternalPort = "internal_port"
	FieldKubeOVNIptablesDnatReady        = "ready"
	FieldKubeOVNIptablesDnatStatusV4IP   = "status_v4_ip"
	FieldKubeOVNIptablesDnatStatusV6IP   = "status_v6_ip"
	FieldKubeOVNIptablesDnatStatusNat    = "status_nat_gw_dp"
	FieldKubeOVNIptablesDnatStatusProto  = "status_protocol"
	FieldKubeOVNIptablesDnatStatusIntIP  = "status_internal_ip"
	FieldKubeOVNIptablesDnatStatusIntP   = "status_internal_port"
	FieldKubeOVNIptablesDnatStatusExtP   = "status_external_port"
)
