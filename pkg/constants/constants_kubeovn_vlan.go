package constants

const (
	ResourceTypeKubeOVNVlan = "harvester_kubeovn_vlan"

	FieldKubeOVNVlanID             = "vlan_id"
	FieldKubeOVNVlanProvider       = "network_provider"
	FieldKubeOVNVlanStatusSubnets  = "status_subnets"
	FieldKubeOVNVlanStatusConflict = "status_conflict"
)

// Bounds of spec.id in the kube-ovn Vlan CRD (0 means untagged). kube-ovn
// does not export them.
const (
	KubeOVNVlanIDMin = 0
	KubeOVNVlanIDMax = 4095
)
