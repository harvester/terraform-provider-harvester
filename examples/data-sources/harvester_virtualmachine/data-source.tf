data "harvester_virtualmachine" "ubuntu20" {
  name      = "ubuntu20"
  namespace = "default"
}

data "harvester_virtualmachine" "opensuse154" {
  name      = "opensuse154"
  namespace = "default"
}