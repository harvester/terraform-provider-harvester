package constants

const (
	DataSourceTypePCIDevice      = "harvester_pcidevice"
	DataSourceTypePCIDeviceClaim = "harvester_pcideviceclaim"

	// PCIDevice data source fields (from CRD status)
	FieldPCIDeviceAddress           = "address"
	FieldPCIDeviceNodeName          = "node_name"
	FieldPCIDeviceVendorID          = "vendor_id"
	FieldPCIDeviceDeviceID          = "device_id"
	FieldPCIDeviceClassID           = "class_id"
	FieldPCIDeviceDeviceDescription = "device_description"
	FieldPCIDeviceIOMMUGroup        = "iommu_group"
	FieldPCIDeviceKernelDriver      = "kernel_driver_in_use"
	FieldPCIDeviceResourceName      = "resource_name"

	// PCIDeviceClaim data source fields
	FieldPCIDeviceClaimNodeName = "node_name"
	FieldPCIDeviceClaimAddress  = "address"

	// VM pci_device block (maps to KubeVirt hostDevices)
	FieldVirtualMachinePCIDevice           = "pci_device"
	FieldVirtualMachinePCIDeviceName       = "name"
	FieldVirtualMachinePCIDeviceDeviceName = "device_name"
)
