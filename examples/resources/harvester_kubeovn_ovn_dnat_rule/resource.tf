resource "harvester_kubeovn_ovn_dnat_rule" "example" {
  name          = "example-ovn-dnat"
  ovn_eip       = "example-ovn-eip"
  internal_port = "8080"
  external_port = "80"
  protocol      = "tcp"
}
