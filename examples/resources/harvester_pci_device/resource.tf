# Example: Attach PCI devices to a Virtual Machine

# First, create a virtual machine
resource "harvester_virtualmachine" "example" {
  name        = "example-vm-with-pci"
  namespace   = "default"
  description = "VM with PCI device passthrough"

  cpu    = 2
  memory = "4Gi"

  # IMPORTANT: The VM must be scheduled on a specific node where PCI devices are available
  # This is handled by the harvester_pci_device resource which sets node_name
  run_strategy = "RerunOnFailure"
  hostname     = "example-vm-pci"

  network_interface {
    name         = "nic-1"
    network_name = "vlan1"
  }

  disk {
    name        = "disk-1"
    type        = "disk"
    size        = "20Gi"
    bus         = "virtio"
    boot_order  = 1
    image       = "harvester-public/image-ubuntu20.04"
    auto_delete = true
  }
}

# Attach PCI devices to the VM
# Note: The PCI devices must be enabled for passthrough in Harvester UI first
resource "harvester_pci_device" "example" {
  name      = "example-pci-device-claim"
  namespace = "default"

  # The VM to attach PCI devices to
  vm_name = "${harvester_virtualmachine.example.namespace}/${harvester_virtualmachine.example.name}"

  # REQUIRED: The node where the VM must be deployed
  # This ensures the VM runs on the correct node where PCI devices are available
  # This prevents scheduling issues when multiple nodes have the same PCI device type
  node_name = "node-01"

  # List of PCI addresses to attach
  # Format: "0000:XX:YY.Z" (e.g., "0000:01:00.0")
  # The PCI devices must be enabled for passthrough in Harvester before they can be attached
  pci_addresses = [
    "0000:01:00.0", # Example: NVIDIA GPU
    "0000:01:00.1", # Example: Additional PCI device
  ]

  labels = {
    environment = "production"
    managed-by  = "terraform"
  }
}

# Example: Multiple PCI devices from different nodes
# Note: Each harvester_pci_device resource can only attach devices from one node
# If you need devices from multiple nodes, you need multiple resources
resource "harvester_pci_device" "example_node2" {
  name      = "example-pci-device-claim-node2"
  namespace = "default"

  vm_name   = "${harvester_virtualmachine.example.namespace}/${harvester_virtualmachine.example.name}"
  node_name = "node-02"

  pci_addresses = [
    "0000:02:00.0",
  ]
}

