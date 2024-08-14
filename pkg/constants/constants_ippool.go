package constants

const (
	ResourceTypeIPPool = "harvester_ippool"

	FieldIPPoolDescription = "description"
)

const (
	SubresourceTypeIPPoolRange = "range"

	FieldRangeStart   = "start"
	FieldRangeEnd     = "end"
	FieldRangeSubnet  = "subnet"
	FieldRangeGateway = "gateway"
)

const (
	SubresourceTypeIPPoolSelector = "selector"

	FieldSelectorPriority = "priority"
	FieldSelectorNetwork  = "network"
)

const (
	SubresourceTypeIPPoolSelectorScope = "scope"

	FieldScopeProject      = "project"
	FieldScopeNamespace    = "namespace"
	FieldScopeGuestCluster = "guest_cluster"
)
