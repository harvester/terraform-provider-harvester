package constants

const (
	ResourceTypeSRIOVNetworkDevice = "harvester_sriov_network_device"
	ResourceTypeSRIOVGPUDevice     = "harvester_sriov_gpu_device"

	FieldSRIOVNetworkDeviceNumVFs = "virtual_functions"

	FieldSRIOVNetworkDeviceEnabled       = "enabled"
	FieldSRIOVNetworkDeviceVFAddresses   = "vf_addresses"
	FieldSRIOVNetworkDeviceVFDeviceNames = "vf_device_names"

	FieldSRIOVGPUDeviceEnabled       = "enabled"
	FieldSRIOVGPUDeviceVFAddresses   = "vf_addresses"
	FieldSRIOVGPUDeviceVFDeviceNames = "vf_device_names"
)
