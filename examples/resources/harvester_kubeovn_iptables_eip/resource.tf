# The address is taken from one of the external subnets of the NAT gateway.
# v4_ip is optional: kube-ovn allocates a free address when it is omitted.
resource "harvester_kubeovn_iptables_eip" "example" {
  name            = "my-eip"
  nat_gw_dp       = harvester_kubeovn_vpc_nat_gateway.example.name
  external_subnet = harvester_kubeovn_subnet.external.name
  v4_ip           = "192.168.100.10"
}
