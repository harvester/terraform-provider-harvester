resource "harvester_kubeovn_ovn_eip" "example" {
  name            = "example-ovn-eip"
  external_subnet = "ovn-vpc-external-network"
  type            = "nat"
}
