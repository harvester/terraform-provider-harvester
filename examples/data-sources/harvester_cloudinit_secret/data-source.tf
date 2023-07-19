data "harvester_cloudinit_secret" "cloud-config-opensuse154" {
  name      = "cloud-config-opensuse154"
  namespace = "default"
}

data "harvester_cloudinit_secret" "cloud-config-ubuntu20" {
  name      = "cloud-config-ubuntu20"
  namespace = "default"
}