resource "harvester_sriov_network_device" "node3-eno50-root" {
  name              = "node3-eno50" # name of the root PCI device to configure
  virtual_functions = 2             # desired number of virtual functions
}
