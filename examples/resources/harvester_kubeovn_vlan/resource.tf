resource "harvester_kubeovn_provider_network" "example" {
  name              = "example-provider-net"
  default_interface = "ens4"
}

# VLAN 100 on the provider network. A kube-ovn subnet joins it by setting
# its vlan attribute to the name of this VLAN.
resource "harvester_kubeovn_vlan" "example" {
  name             = "example-vlan"
  vlan_id          = 100
  network_provider = harvester_kubeovn_provider_network.example.name
}
