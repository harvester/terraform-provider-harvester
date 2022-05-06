package constants

const (
	ResourceTypeVirtualMachine = "harvester_virtualmachine"

	FieldVirtualMachineMachineType      = "machine_type"
	FieldVirtualMachineHostname         = "hostname"
	FieldVirtualMachineStart            = "start"
	FieldVirtualMachineRunStrategy      = "run_strategy"
	FieldVirtualMachineCPU              = "cpu"
	FieldVirtualMachineMemory           = "memory"
	FieldVirtualMachineSSHKeys          = "ssh_keys"
	FieldVirtualMachineCloudInit        = "cloudinit"
	FieldVirtualMachineDisk             = "disk"
	FieldVirtualMachineNetworkInterface = "network_interface"
	FieldVirtualMachineInstanceNodeName = "node_name"

	StateVirtualMachineStarting = "starting"
	StateVirtualMachineRunning  = "running"
	StateVirtualMachineStopping = "stopping"
	StateVirtualMachineStopped  = "stopped"
	StateVirtualMachineError    = "error"
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
	FiledNetworkInterfaceName          = "name"
	FiledNetworkInterfaceType          = "type"
	FiledNetworkInterfaceModel         = "model"
	FiledNetworkInterfaceMACAddress    = "mac_address"
	FiledNetworkInterfaceIPAddress     = "ip_address"
	FiledNetworkInterfaceInterfaceName = "interface_name"
	FiledNetworkInterfaceNetworkName   = "network_name"
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
