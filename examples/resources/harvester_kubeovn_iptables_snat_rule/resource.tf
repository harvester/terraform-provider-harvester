resource "harvester_kubeovn_iptables_snat_rule" "example" {
  name          = "my-snat-rule"
  eip           = harvester_kubeovn_iptables_eip.example.name
  internal_cidr = "10.0.0.0/24"
}
