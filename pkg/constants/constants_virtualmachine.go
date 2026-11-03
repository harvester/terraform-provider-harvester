package constants

const (
	ResourceTypeVirtualMachine = "harvester_virtualmachine"

	FieldRequestsCPU                              = "cpu"
	FieldRequestsMemory                           = "memory"
	FieldVirtualMachineBlockMultiQueue            = "block_device_multiqueue"
	FieldVirtualMachineCPU                        = "cpu"
	FieldVirtualMachineCPUModel                   = "cpu_model"
	FieldVirtualMachineCPUPinning                 = "cpu_pinning"
	FieldVirtualMachineCloudInit                  = "cloudinit"
	FieldVirtualMachineCreateInitialSnapshot      = "create_initial_snapshot"
	FieldVirtualMachineDisk                       = "disk"
	FieldVirtualMachineEFI                        = "efi"
	FieldVirtualMachineHostDevice                 = "host_device"
	FieldVirtualMachineHostname                   = "hostname"
	FieldVirtualMachineIOThreadsCount             = "io_threads_count"
	FieldVirtualMachineIOThreadsPolicy            = "io_threads_policy"
	FieldVirtualMachineInput                      = "input"
	FieldVirtualMachineInstanceNodeName           = "node_name"
	FieldVirtualMachineIsolateEmulatorThread      = "isolate_emulator_thread"
	FieldVirtualMachineMachineType                = "machine_type"
	FieldVirtualMachineMemory                     = "memory"
	FieldVirtualMachineNetworkInterface           = "network_interface"
	FieldVirtualMachineNetworkInterfaceMultiqueue = "network_interface_multiqueue"
	FieldVirtualMachineNodeSelector               = "node_selector"
	FieldVirtualMachineRequests                   = "requests"
	FieldVirtualMachineReservedMemory             = "reserved_memory"
	FieldVirtualMachineRestartAfterUpdate         = "restart_after_update"
	FieldVirtualMachineRunStrategy                = "run_strategy"
	FieldVirtualMachineSSHKeys                    = "ssh_keys"
	FieldVirtualMachineSecureBoot                 = "secure_boot"
	FieldVirtualMachineStart                      = "start"
	FieldVirtualMachineTPM                        = "tpm"

	// Node Affinity - Controls VM scheduling based on node labels
	// Reference: https://docs.harvesterhci.io/v1.7/vm/index/#node-scheduling
	FieldVirtualMachineNodeAffinity = "node_affinity"
	FieldNodeAffinityRequired       = "required"  // requiredDuringSchedulingIgnoredDuringExecution
	FieldNodeAffinityPreferred      = "preferred" // preferredDuringSchedulingIgnoredDuringExecution
	FieldNodeSelectorTerm           = "node_selector_term"
	FieldMatchExpressions           = "match_expressions" // Match by labels
	FieldMatchFields                = "match_fields"      // Match by node fields
	FieldExpressionKey              = "key"
	FieldExpressionOperator         = "operator" // In, NotIn, Exists, DoesNotExist, Gt, Lt
	FieldExpressionValues           = "values"
	FieldPreferredWeight            = "weight"     // 1-100, higher means more preferred
	FieldPreferredPreference        = "preference" // Node selector term for preferred scheduling

	// Pod Affinity/Anti-Affinity - Controls VM co-location with other pods
	// Reference: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#inter-pod-affinity-and-anti-affinity
	FieldVirtualMachinePodAffinity     = "pod_affinity"      // Co-locate VMs with matching pods
	FieldVirtualMachinePodAntiAffinity = "pod_anti_affinity" // Separate VMs from matching pods
	FieldPodAffinityRequired           = "required"          // requiredDuringSchedulingIgnoredDuringExecution
	FieldPodAffinityPreferred          = "preferred"         // preferredDuringSchedulingIgnoredDuringExecution
	FieldLabelSelector                 = "label_selector"    // Select pods by labels
	FieldMatchLabels                   = "match_labels"      // Exact label matching
	FieldNamespaces                    = "namespaces"        // Limit to specific namespaces
	FieldNamespaceSelector             = "namespace_selector"
	FieldTopologyKey                   = "topology_key" // e.g., kubernetes.io/hostname
	FieldPodAffinityTerm               = "pod_affinity_term"

	StateVirtualMachineStarting = "Starting"
	StateVirtualMachineRunning  = "Running"
	StateVirtualMachineStopping = "Stopping"
	StateVirtualMachineStopped  = "Off"
)

const (
	ResourceVirtualMachine = "virtualmachines"
	SubresourceRestart     = "restart"
)

const (
	FieldCloudInitType                  = "type"
	FieldCloudInitNetworkData           = "network_data"
	FieldCloudInitNetworkDataBase64     = "network_data_base64"
	FieldCloudInitNetworkDataSecretName = "network_data_secret_name"
	FieldCloudInitUserData              = "user_data"
	FieldCloudInitUserDataBase64        = "user_data_base64"
	FieldCloudInitUserDataSecretName    = "user_data_secret_name"
)

// NetworkInterfaceBindingManagedTap is the KubeVirt network binding plugin
// that the Harvester VM webhook puts in place of the bridge binding when the
// interface is on a KubeOVN subnet with DHCP enabled. Same value as
// ManagedTapBindingName in github.com/harvester/harvester/pkg/util, which the
// Harvester version used by the provider does not have yet.
const NetworkInterfaceBindingManagedTap = "managedtap"

const (
	FieldNetworkInterfaceName          = "name"
	FieldNetworkInterfaceType          = "type"
	FieldNetworkInterfaceModel         = "model"
	FieldNetworkInterfaceMACAddress    = "mac_address"
	FieldNetworkInterfaceIPAddress     = "ip_address"
	FieldNetworkInterfaceInterfaceName = "interface_name"
	FieldNetworkInterfaceWaitForLease  = "wait_for_lease"
	FieldNetworkInterfaceNetworkName   = "network_name"
	FieldNetworkInterfaceBootOrder     = "boot_order"
)

const (
	FieldDiskName               = "name"
	FieldDiskType               = "type"
	FieldDiskSize               = "size"
	FieldDiskBus                = "bus"
	FieldDiskCacheMode          = "cache_mode"
	FieldDiskBootOrder          = "boot_order"
	FieldDiskExistingVolumeName = "existing_volume_name"
	FieldDiskContainerImageName = "container_image_name"
	FieldDiskHotPlug            = "hot_plug"
	FieldDiskAutoDelete         = "auto_delete"
	FieldDiskVolumeName         = "volume_name"
	FieldDiskDedicatedIOThread  = "dedicated_io_thread"

	AnnotationDiskAutoDelete = "terraform-provider-harvester-auto-delete"
)

const (
	FieldInputName = "name"
	FieldInputType = "type"
	FieldInputBus  = "bus"
)

const (
	FieldTPMName = "name"
)

const (
	LabelSSHUsername = "ssh-user"
)

const (
	DiskCacheModeNone         = "none"
	DiskCacheModeWriteBack    = "writeback"
	DiskCacheModeWriteThrough = "writethrough"
)

const (
	FieldHostDeviceName       = "name"
	FieldHostDeviceDeviceName = "device_name"
)
