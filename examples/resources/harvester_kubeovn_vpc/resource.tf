resource "harvester_kubeovn_vpc" "example" {
  name = "my-vpc"

  namespaces = ["default", "my-namespace"]

  static_routes {
    cidr        = "192.0.2.0/24"
    next_hop_ip = "10.0.0.254"
  }

  policy_routes {
    priority = 10
    match    = "ip4.src == 10.0.0.0/24"
    action   = "allow"
  }
}
