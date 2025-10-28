package constants

const (
	NamespaceDefault         = "default"
	NamespaceHarvesterSystem = "harvester-system"

	FieldProviderBootstrap   = "bootstrap"
	FieldProviderKubeConfig  = "kubeconfig"
	FieldProviderKubeContext = "kubecontext"

	FieldCommonName        = "name"
	FieldCommonNamespace   = "namespace"
	FieldCommonTags        = "tags"
	FieldCommonLabels      = "labels"
	FieldCommonDescription = "description"
	FieldCommonState       = "state"
	FieldCommonMessage     = "message"

	StateCommonActive  = "Active"
	StateCommonReady   = "Ready"
	StateCommonRemoved = "Removed"
	StateCommonError   = "Error"
	StateCommonFailed  = "Failed"
	StateCommonUnknown = "Unknown"
)
