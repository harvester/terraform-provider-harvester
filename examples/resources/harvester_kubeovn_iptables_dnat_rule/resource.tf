resource "harvester_kubeovn_iptables_dnat_rule" "example" {
  name          = "my-dnat-rule"
  eip           = harvester_kubeovn_iptables_eip.example.name
  external_port = "8080"
  protocol      = "tcp"
  internal_ip   = "10.0.0.100"
  internal_port = "80"
}
