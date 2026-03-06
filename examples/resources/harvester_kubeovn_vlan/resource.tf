resource "harvester_kubeovn_vlan" "example" {
  name             = "example-vlan"
  vlan_id          = 100
  network_provider = "provider-net1"
}
