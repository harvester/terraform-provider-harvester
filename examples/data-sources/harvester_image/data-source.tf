data "harvester_image" "ubuntu20" {
  name      = "ubuntu20"
  namespace = "harvester-public"
}

data "harvester_image" "ubuntu1804" {
  namespace    = "harvester-public"
  display_name = "ubuntu-18.04-minimal-cloudimg-amd64.img"
}