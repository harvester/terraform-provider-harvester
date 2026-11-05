# The VMs of the subnet reach external networks with the address of the EIP.
resource "harvester_kubeovn_iptables_snat_rule" "example" {
  name          = "my-snat-rule"
  eip           = harvester_kubeovn_iptables_eip.example.name
  internal_cidr = harvester_kubeovn_subnet.example.cidr_block
}
