// Package constants defines constants used by the harvester_pci_device resource.
package constants

const (
	// ResourceTypePCIDevice is the Terraform resource type name for harvester_pci_device.
	// This resource manages PCI device passthrough to VMs using Harvester's PCIDeviceClaim CRD.
	ResourceTypePCIDevice = "harvester_pci_device"

	// FieldPCIDeviceVMName is the field name for the VM to attach PCI devices to (required).
	// Format: "namespace/name" or "name" (if in default namespace).
	FieldPCIDeviceVMName = "vm_name"

	// FieldPCIDeviceNodeName is the field name for the node where the VM must be deployed (required).
	// This ensures the VM runs on a specific node where the PCI devices are available.
	FieldPCIDeviceNodeName = "node_name"

	// FieldPCIDevicePCIAddresses is the field name for the list of PCI addresses (required).
	// Each address should be in the format "0000:XX:YY.Z" (e.g., "0000:01:00.0").
	FieldPCIDevicePCIAddresses = "pci_addresses"

	// FieldPCIDeviceLabels are labels to apply to the PCIDeviceClaim resource (optional).
	FieldPCIDeviceLabels = "labels"
)
