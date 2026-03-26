resource "harvester_sriov_network_device" "eno50-root" {
  name              = "hp-49-tink-system-eno50" # name of the root PCI device to configure
  virtual_functions = 2                         # desired number of virtual functions
}
