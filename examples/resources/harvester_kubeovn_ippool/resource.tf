resource "harvester_kubeovn_ippool" "example" {
  name   = "example-ippool"
  subnet = "example-subnet"

  ips = [
    "10.0.0.10",
    "10.0.0.20..10.0.0.30",
  ]

  namespaces = [
    "default",
  ]
}
