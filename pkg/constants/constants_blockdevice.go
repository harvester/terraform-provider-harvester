package constants

const (
	ResourceTypeBlockDevice = "harvester_blockdevice"

	NamespaceLonghornSystem = "longhorn-system"

	FieldBlockDeviceNodeName       = "node_name"
	FieldBlockDeviceDevPath        = "dev_path"
	FieldBlockDeviceProvision      = "provision"
	FieldBlockDeviceForceFormatted = "force_formatted"
	FieldBlockDeviceDeviceTags     = "device_tags"

	FieldBlockDeviceProvisioner              = "provisioner"
	FieldBlockDeviceProvisionerLonghorn      = "longhorn"
	FieldBlockDeviceProvisionerLonghornEV    = "engine_version"
	FieldBlockDeviceProvisionerLonghornDD    = "disk_driver"
	FieldBlockDeviceProvisionerLVM           = "lvm"
	FieldBlockDeviceProvisionerLVMVGName     = "vg_name"
	FieldBlockDeviceProvisionerLVMParameters = "parameters"

	FieldBlockDeviceProvisionPhase = "provision_phase"
	FieldBlockDeviceDeviceStatus   = "device_status"

	FieldBlockDeviceStatusDevPath           = "dev_path"
	FieldBlockDeviceStatusParentDevice      = "parent_device"
	FieldBlockDeviceStatusPartitioned       = "partitioned"
	FieldBlockDeviceStatusCapacitySizeBytes = "capacity_size_bytes"
	FieldBlockDeviceStatusDeviceType        = "device_type"
	FieldBlockDeviceStatusDriveType         = "drive_type"
	FieldBlockDeviceStatusStorageController = "storage_controller"
	FieldBlockDeviceStatusVendor            = "vendor"
	FieldBlockDeviceStatusModel             = "model"
	FieldBlockDeviceStatusSerialNumber      = "serial_number"
	FieldBlockDeviceStatusWWN               = "wwn"
	FieldBlockDeviceStatusBusPath           = "bus_path"
	FieldBlockDeviceStatusFSType            = "filesystem_type"
	FieldBlockDeviceStatusMountPoint        = "mount_point"
	FieldBlockDeviceStatusIsReadOnly        = "is_read_only"
	FieldBlockDeviceStatusIsRemovable       = "is_removable"
)
