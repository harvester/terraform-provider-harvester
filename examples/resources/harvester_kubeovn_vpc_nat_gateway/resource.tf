resource "harvester_kubeovn_vpc_nat_gateway" "example" {
  name   = "my-nat-gateway"
  vpc    = harvester_kubeovn_vpc.example.name
  subnet = harvester_kubeovn_subnet.example.name

  external_subnets = ["ovn-default"]
}
