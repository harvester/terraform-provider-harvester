resource "harvester_kubeovn_security_group" "example" {
  name                     = "example-sg"
  allow_same_group_traffic = true

  ingress_rules {
    ip_version     = "ipv4"
    protocol       = "tcp"
    priority       = 1
    remote_type    = "address"
    remote_address = "10.0.0.0/24"
    port_range_min = 80
    port_range_max = 80
    policy         = "allow"
  }

  egress_rules {
    ip_version     = "ipv4"
    protocol       = "all"
    remote_type    = "address"
    remote_address = "0.0.0.0/0"
    policy         = "allow"
  }
}
