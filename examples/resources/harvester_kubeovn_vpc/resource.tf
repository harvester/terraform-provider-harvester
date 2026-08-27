resource "harvester_kubeovn_vpc" "example" {
  name = "my-vpc"

  namespaces = ["default", "my-namespace"]

  enable_external = false
  enable_bfd      = false

  static_routes {
    cidr        = "10.0.0.0/24"
    next_hop_ip = "10.0.0.1"
  }

  policy_routes {
    priority = 10
    match    = "ip4.src == 10.0.0.0/24"
    action   = "allow"
  }
}
