# Bridge a spare host interface into kube-ovn. Every node must have the
# interface, unless it is excluded or uses one of custom_interfaces.
resource "harvester_kubeovn_provider_network" "example" {
  name              = "example-provider-net"
  default_interface = "ens4"

  # node3 has its spare port on another interface name.
  custom_interfaces {
    interface_name = "ens5"
    nodes          = ["node3"]
  }

  # node4 has no spare port: kube-ovn leaves it out.
  exclude_nodes = ["node4"]
}
