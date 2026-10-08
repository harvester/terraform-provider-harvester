resource "harvester_kubeovn_provider_network" "example" {
  name              = "example-provider-net"
  default_interface = "eth0"

  custom_interfaces {
    interface_name = "eth1"
    nodes          = ["node1", "node2"]
  }

  exclude_nodes      = ["node3"]
  exchange_link_name = false
}
