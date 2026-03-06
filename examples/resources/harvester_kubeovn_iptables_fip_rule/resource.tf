resource "harvester_kubeovn_iptables_fip_rule" "example" {
  name        = "my-fip-rule"
  eip         = harvester_kubeovn_iptables_eip.example.name
  internal_ip = "10.0.0.100"
}
