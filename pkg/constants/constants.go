package constants

const (
	NamespaceDefault         = "default"
	NamespaceHarvesterSystem = "harvester-system"

	FieldProviderKubeConfig       = "kubeconfig"
	FieldProviderKubeContext      = "kubecontext"
	FieldProviderKubeConfigBase64 = "kubeconfig_base64"

	FieldCommonName        = "name"
	FieldCommonNamespace   = "namespace"
	FieldCommonTags        = "tags"
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
