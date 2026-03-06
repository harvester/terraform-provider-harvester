resource "harvester_kubeovn_ovn_fip" "example" {
  name    = "example-ovn-fip"
  ovn_eip = "example-ovn-eip"
}
