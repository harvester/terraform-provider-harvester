package constants

const (
	ResourceTypeKubeOVNProviderNetwork = "harvester_kubeovn_provider_network"

	FieldKubeOVNProviderNetDefaultInterface    = "default_interface"
	FieldKubeOVNProviderNetCustomInterfaces    = "custom_interfaces"
	FieldKubeOVNProviderNetExcludeNodes        = "exclude_nodes"
	FieldKubeOVNProviderNetExchangeLinkName    = "exchange_link_name"
	FieldKubeOVNProviderNetStatusReady         = "status_ready"
	FieldKubeOVNProviderNetStatusReadyNodes    = "status_ready_nodes"
	FieldKubeOVNProviderNetStatusNotReadyNodes = "status_not_ready_nodes"
	FieldKubeOVNProviderNetStatusVlans         = "status_vlans"

	FieldKubeOVNCustomInterfaceInterface = "interface_name"
	FieldKubeOVNCustomInterfaceNodes     = "nodes"
)
