resource "harvester_clusternetwork" "vlan" {
  lifecycle {
    prevent_destroy = true
  }
  name                 = "vlan"
  enable               = true
  default_physical_nic = "eth0"
}