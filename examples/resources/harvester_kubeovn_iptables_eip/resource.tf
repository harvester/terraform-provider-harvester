resource "harvester_kubeovn_iptables_eip" "example" {
  name      = "my-eip"
  nat_gw_dp = harvester_kubeovn_vpc_nat_gateway.example.name
}
