package constants

const (
	ResourceTypeVLANConfig = "harvester_vlanconfig"

	FieldVLANConfigClusterNetworkName = "cluster_network_name"
	FieldVLANConfigNodeSelector       = "node_selector"
	FieldVLANConfigMatchedNodes       = "matched_nodes"
	FieldVLANConfigUplink             = "uplink"
)

const (
	FieldUplinkNICs         = "nics"
	FieldUplinkBondMode     = "bond_mode"
	FieldUplinkBondMiimon   = "bond_miimon"
	FieldUplinkMTU          = "mtu"
	FieldUplinkTxQLen       = "txq_len"
	FieldUplinkHardwareAddr = "hardware_addr"
)
