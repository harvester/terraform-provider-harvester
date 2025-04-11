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
