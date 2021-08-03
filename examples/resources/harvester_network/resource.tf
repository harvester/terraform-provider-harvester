resource "harvester_network" "vlan1" {
  name    = "vlan1"
  vlan_id = 1
}

resource "harvester_network" "vlan" {
  for_each = toset(["2", "3"])
  name     = "vlan${each.key}"
  vlan_id  = each.key
}