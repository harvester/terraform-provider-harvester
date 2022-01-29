resource "harvester_network" "vlan1" {
  name      = "vlan1"
  namespace = "harvester-public"

  vlan_id = 1

  route_dhcp_server_ip = ""
}

resource "harvester_network" "vlan" {
  for_each  = toset(["2", "3"])
  name      = "vlan${each.key}"
  namespace = "harvester-public"

  vlan_id = each.key

  route_mode    = "manual"
  route_cidr    = "172.16.0.1/24"
  route_gateway = "172.16.0.1"
}