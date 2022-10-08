resource "harvester_vlanconfig" "cluster-vlan-node1" {
  name = "cluster-vlan-node1"

  cluster_network_name = harvester_clusternetwork.cluster-vlan.name

  uplink {
    nics = [
      "eth5",
      "eth6"
    ]

    bond_mode = "active-backup"
    mtu       = 1500
  }

  node_selector = {
    "kubernetes.io/hostname" : "node1"
  }
}