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
	FieldVirtualMachineHyperv                = "hyperv"
	FieldVirtualMachineHypervPassthrough     = "hyperv_passthrough" // #nosec G101
	FieldVirtualMachineClock                 = "clock"

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
	FieldDiskName                 = "name"
	FieldDiskType                 = "type"
	FieldDiskSize                 = "size"
	FieldDiskBus                  = "bus"
	FieldDiskBootOrder            = "boot_order"
	FieldDiskExistingVolumeName   = "existing_volume_name"
	FieldDiskContainerImageName   = "container_image_name"
	FieldDiskHotPlug              = "hot_plug"
	FieldDiskAutoDelete           = "auto_delete"
	FieldDiskVolumeName           = "volume_name"
	FieldDiskSysprepSecretName    = "sysprep_secret_name" // #nosec G101
	FieldDiskSysprepConfigMapName = "sysprep_configmap_name"

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
	FieldHypervRelaxed          = "relaxed"
	FieldHypervVAPIC            = "vapic"
	FieldHypervVPIndex          = "vpindex"
	FieldHypervRuntime          = "runtime"
	FieldHypervSyNIC            = "synic"
	FieldHypervReset            = "reset"
	FieldHypervFrequencies      = "frequencies"
	FieldHypervReenlightenment  = "reenlightenment"
	FieldHypervTLBFlush         = "tlbflush"
	FieldHypervIPI              = "ipi"
	FieldHypervEVMCS            = "evmcs"
	FieldHypervSpinlocks        = "spinlocks"
	FieldHypervSpinlocksRetries = "spinlocks_retries"
	FieldHypervSyNICTimer       = "synictimer"
	FieldHypervSyNICTimerDirect = "synictimer_direct"
	FieldHypervVendorID         = "vendorid"
	FieldHypervVendorIDValue    = "vendorid_value"
)

const (
	FieldClockTimezone         = "timezone"
	FieldClockUTCOffsetSeconds = "utc_offset_seconds"
	FieldClockTimer            = "timer"
	FieldTimerHPET             = "hpet"
	FieldTimerKVM              = "kvm"
	FieldTimerPIT              = "pit"
	FieldTimerRTC              = "rtc"
	FieldTimerHyperv           = "hyperv"
	FieldTimerEnabled          = "enabled"
	FieldTimerTickPolicy       = "tick_policy"
	FieldTimerTrack            = "track"
)
