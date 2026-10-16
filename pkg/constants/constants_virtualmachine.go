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
	FieldVirtualMachineRequests              = "requests"
	FieldRequestsCPU                         = "cpu"
	FieldRequestsMemory                      = "memory"
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
	FieldVirtualMachineHostDevice            = "host_device"

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
