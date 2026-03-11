# Look up a PCI device by name to discover its properties
data "harvester_pcidevice" "gpu" {
  name = "node1-000001000"
}

output "gpu_device" {
  value = {
    address       = data.harvester_pcidevice.gpu.address
    node_name     = data.harvester_pcidevice.gpu.node_name
    vendor_id     = data.harvester_pcidevice.gpu.vendor_id
    device_id     = data.harvester_pcidevice.gpu.device_id
    description   = data.harvester_pcidevice.gpu.device_description
    resource_name = data.harvester_pcidevice.gpu.resource_name
  }
}

# Use the discovered device in a VM via hostDevices passthrough
resource "harvester_virtualmachine" "gpu_vm" {
  name      = "gpu-workload"
  namespace = "default"
  # ... other VM config ...

  pci_device {
    name        = "gpu0"
    device_name = data.harvester_pcidevice.gpu.resource_name
  }
}
