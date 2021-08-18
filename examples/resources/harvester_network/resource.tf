resource "harvester_network" "vlan1" {
  name      = "vlan1"
  namespace = "harvester-public"

  vlan_id = 1
}

resource "harvester_network" "vlan" {
  for_each  = toset(["2", "3"])
  name      = "vlan${each.key}"
  namespace = "harvester-public"

  vlan_id = each.key
}