resource "harvester_kubeovn_vpc_egress_gateway" "example" {
  name            = "test-egress-gw"
  namespace       = "default"
  vpc             = "test-vpc"
  replicas        = 1
  external_subnet = "ovn-vpc-external-network"
  traffic_policy  = "Cluster"

  bfd {
    enabled = false
  }

  policy {
    snat      = true
    ip_blocks = ["0.0.0.0/0"]
  }
}
