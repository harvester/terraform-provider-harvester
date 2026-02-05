package constants

const (
	ResourceTypeVirtualMachine = "harvester_virtualmachine"

	FieldVirtualMachineMachineType           = "machine_type"
	FieldVirtualMachineHostname              = "hostname"
	FieldVirtualMachineReservedMemory        = "reserved_memory"
	FieldVirtualMachineRestartAfterUpdate    = "restart_after_update"
	FieldVirtualMachineStart                 = "start"
	FieldVirtualMachineRunStrategy           = "run_strategy"
	FieldVirtualMachineCPU                   = "cpu"
	FieldVirtualMachineCPUModel              = "cpu_model"
	FieldVirtualMachineMemory                = "memory"
	FieldVirtualMachineSSHKeys               = "ssh_keys"
	FieldVirtualMachineCloudInit             = "cloudinit"
	FieldVirtualMachineDisk                  = "disk"
	FieldVirtualMachineNetworkInterface      = "network_interface"
	FieldVirtualMachineInput                 = "input"
	FieldVirtualMachineTPM                   = "tpm"
	FieldVirtualMachineInstanceNodeName      = "node_name"
	FieldVirtualMachineEFI                   = "efi"
	FieldVirtualMachineSecureBoot            = "secure_boot"
	FieldVirtualMachineCPUPinning            = "cpu_pinning"
	FieldVirtualMachineIsolateEmulatorThread = "isolate_emulator_thread"
	FieldVirtualMachineNodeSelector          = "node_selector"
	FieldVirtualMachineCreateInitialSnapshot = "create_initial_snapshot"

	// Node Affinity - Controls VM scheduling based on node labels
	// Reference: https://docs.harvesterhci.io/v1.7/vm/index/#node-scheduling
	FieldVirtualMachineNodeAffinity = "node_affinity"
	FieldNodeAffinityRequired       = "required"  // requiredDuringSchedulingIgnoredDuringExecution
	FieldNodeAffinityPreferred      = "preferred" // preferredDuringSchedulingIgnoredDuringExecution
	FieldNodeSelectorTerm           = "node_selector_term"
	FieldMatchExpressions           = "match_expressions" // Match by node labels
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
	FieldDiskBootOrder          = "boot_order"
	FieldDiskExistingVolumeName = "existing_volume_name"
	FieldDiskContainerImageName = "container_image_name"
	FieldDiskHotPlug            = "hot_plug"
	FieldDiskAutoDelete         = "auto_delete"
	FieldDiskVolumeName         = "volume_name"

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
