# Example: Attach PCI devices to a VM via passthrough
# PCI devices must be enabled for passthrough in Harvester UI first.

resource "harvester_pci_device" "gpu" {
  name      = "gpu-passthrough"
  namespace = "default"

  vm_name   = "default/my-vm"
  node_name = "node-01"

  # PCI addresses in format "0000:XX:YY.Z"
  pci_addresses = [
    "0000:01:00.0",
    "0000:01:00.1",
  ]

  labels = {
    device-type = "gpu"
  }
}
