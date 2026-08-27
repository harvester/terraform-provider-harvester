package constants

const (
	ResourceTypeVirtualMachine = "harvester_virtualmachine"

	FieldRequestsCPU                              = "cpu"
	FieldRequestsMemory                           = "memory"
	FieldVirtualMachineBlockMultiQueue            = "block_multi_queue"
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
