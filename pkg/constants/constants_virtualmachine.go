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
	FieldVirtualMachineFeatures              = "features"

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

const (
	FieldFeatureACPI = "acpi" // enabled/disabled

	FieldFeatureAPIC               = "apic"
	FieldFeatureAPICEnabled        = "enabled"
	FieldFeatureAPICEndOfInterrupt = "end_of_interrupt"

	FieldFeatureHyperV                = "hyperv"
	FieldFeatureHyperVEVMCS           = "evmcs"           //enabled/disabled
	FieldFeatureHyperVFrequencies     = "frequencies"     //enabled/disabled
	FieldFeatureHyperVIPI             = "ipi"             //enabled/disabled
	FieldFeatureHyperVReenlightenment = "reenlightenment" //enabled/disabled
	FieldFeatureHyperVRelaxed         = "relaxed"         //enabled/disabled
	FieldFeatureHyperVReset           = "reset"           //enabled/disabled
	FieldFeatureHyperVRuntime         = "runtime"         //enabled/disabled

	FieldFeatureHyperVSpinlocks        = "spinlocks"
	FieldFeatureHyperVSpinlocksEnabled = "enabled"
	FieldFeatureHyperVSpinlocksRetries = "retries"

	FieldFeatureHyperVSyNIC = "synic" //enabled/disabled

	FieldFeatureHyperVSyNICTimer        = "synic_timer"
	FieldFeatureHyperVSyNICTimerEnabled = "enabled"
	FieldFeatureHyperVSyNICTimerDirect  = "direct"

	FieldFeatureHyperVTLBFlush = "tlb_flush" //enabled/disabled
	FieldFeatureHyperVVAPIC    = "vapic"     //enabled/disabled

	FieldFeatureHyperVVendorId = "vendorid" // string, enabled if given

	FieldFeatureHyperVVPIndex = "vpindex" //enabled/disabled

	FieldFeatureHyperVPassthrough = "hyperv_passthrough" //gosec:disable G101 -- false positive, enable/disable

	FieldFeatureKVM       = "kvm"
	FieldFeatureKVMHidden = "hidden"

	FieldFeaturePVSpinLock = "pvspinlock"
	FieldFeatureSMM        = "smm"
)
