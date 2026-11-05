# The external subnet holds the addresses of the EIPs. kube-ovn refuses an
# external subnet that is not backed by a network attachment definition: on
# Harvester, create the network in the UI (Networks > Virtual Machine
# Networks), here "my-external-network" in namespace "default", and point
# network_provider at it.
resource "harvester_kubeovn_subnet" "external" {
  name             = "my-external-subnet"
  cidr_block       = "192.168.100.0/24"
  gateway          = "192.168.100.1"
  network_provider = "my-external-network.default.ovn"
}

# lan_ip is the address of the gateway in the VPC subnet. Route the external
# traffic of the VPC to it, for example with a static route of the VPC from
# 0.0.0.0/0 to 10.0.0.254.
resource "harvester_kubeovn_vpc_nat_gateway" "example" {
  name   = "my-nat-gateway"
  vpc    = harvester_kubeovn_vpc.example.name
  subnet = harvester_kubeovn_subnet.example.name
  lan_ip = "10.0.0.254"

  external_subnets = [harvester_kubeovn_subnet.external.name]
}
