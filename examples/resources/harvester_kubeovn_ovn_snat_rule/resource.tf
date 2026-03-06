resource "harvester_kubeovn_ovn_snat_rule" "example" {
  name       = "example-ovn-snat"
  ovn_eip    = "example-ovn-eip"
  vpc_subnet = "example-subnet"
}
