resource "harvester_kubeovn_subnet" "example" {
  name = "my-subnet"

  vpc        = harvester_kubeovn_vpc.example.name
  cidr_block = "10.0.0.0/24"
  gateway    = "10.0.0.1"

  exclude_ips  = ["10.0.0.1"]
  protocol     = "IPv4"
  nat_outgoing = true
  gateway_type = "distributed"
  enable_lb    = true
}
