resource "harvester_network" "mgmt-vlan1" {
  name      = "mgmt-vlan1"
  namespace = "harvester-public"

  vlan_id = 1

  route_mode           = "auto"
  route_dhcp_server_ip = ""

  cluster_network_name = data.harvester_clusternetwork.mgmt.name
}

resource "harvester_network" "cluster-vlan1" {
  name      = "cluster-vlan1"
  namespace = "harvester-public"

  vlan_id = 1

  route_mode           = "auto"
  route_dhcp_server_ip = ""

  cluster_network_name = harvester_clusternetwork.cluster-vlan.name
  depends_on = [
    harvester_vlanconfig.cluster-vlan-node1
  ]
}

resource "harvester_network" "cluster-vlan" {
  for_each  = toset(["2", "3"])
  name      = "cluster-vlan${each.key}"
  namespace = "harvester-public"

  vlan_id = each.key

  route_mode    = "manual"
  route_cidr    = "172.16.0.1/24"
  route_gateway = "172.16.0.1"

  cluster_network_name = harvester_clusternetwork.cluster-vlan.name
  depends_on = [
    harvester_vlanconfig.cluster-vlan-node1
  ]
}