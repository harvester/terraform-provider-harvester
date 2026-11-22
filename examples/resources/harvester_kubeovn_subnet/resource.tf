# The subnet is attached to an overlay network, created in the Harvester UI
# (Networks > Virtual Machine Networks, type OverlayNetwork). Its provider is
# <network name>.<network namespace>.ovn.
resource "harvester_kubeovn_subnet" "example" {
  name = "my-subnet"

  vpc              = harvester_kubeovn_vpc.example.name
  cidr_block       = "10.0.0.0/24"
  network_provider = "my-overlay-network.default.ovn"

  # The gateway defaults to the first address of cidr_block, and kube-ovn
  # always reserves it: it does not need to be listed in exclude_ips.
  exclude_ips  = ["10.0.0.2..10.0.0.10"]
  nat_outgoing = true
  gateway_type = "distributed"
  enable_lb    = true
}
