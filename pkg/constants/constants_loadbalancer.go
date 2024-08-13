package constants

const (
	ResourceTypeLoadBalancer = "harvester_loadbalancer"

	FieldLoadBalancerDescription           = "description"
	FieldLoadBalancerWorkloadType          = "workload_type"
	FieldLoadBalancerIPAM                  = "ipam"
	FieldLoadBalancerIPPool                = "ippool"
	FieldLoadBalancerBackendServerSelector = "backend_server_selector"
)

const (
	SubresourceTypeLoadBalancerListener = "listener"

	FieldListenerName        = "name"
	FieldListenerPort        = "port"
	FieldListenerProtocol    = "protocol"
	FieldListenerBackendPort = "backend_port"
)

const (
	SubresourceTypeLoadBalancerHealthCheck = "healthcheck"

	FieldHealthcheckPort             = "port"
	FieldHealthcheckSuccessThreshold = "success_threshold"
	FieldHealthcheckFailureThreshold = "failure_threshold"
	FieldHealthcheckPeriodSeconds    = "period_seconds"
	FieldHealthcheckTimeoutSeconds   = "timeout_seconds"
)

const (
	LoadBalancerWorkloadTypeVM      = "vm"
	LoadBalancerWorkloadTypeCluster = "cluster"
)

const (
	LoadBalancerIPAMPool = "pool"
	LoadBalancerIPAMDHCP = "dhcp"
)
