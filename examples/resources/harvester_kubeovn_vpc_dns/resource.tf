resource "harvester_kubeovn_vpc_dns" "example" {
  name     = "example-vpc-dns"
  replicas = 1
  vpc      = "example-vpc"
  subnet   = "example-subnet"
}
