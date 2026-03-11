# Look up an existing PCI device claim
data "harvester_pcideviceclaim" "gpu" {
  name = "node1-000001000"
}

output "gpu_claim" {
  value = {
    node_name = data.harvester_pcideviceclaim.gpu.node_name
    address   = data.harvester_pcideviceclaim.gpu.address
  }
}
