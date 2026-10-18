# Port 8022 of the EIP is forwarded to the SSH port of a VM of the subnet.
resource "harvester_kubeovn_iptables_dnat_rule" "example" {
  name          = "my-dnat-rule"
  eip           = harvester_kubeovn_iptables_eip.example.name
  external_port = "8022"
  protocol      = "tcp"
  internal_ip   = "10.0.0.100"
  internal_port = "22"
}
