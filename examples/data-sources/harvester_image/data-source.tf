data "harvester_image" "ubuntu20" {
  name      = "ubuntu20"
  namespace = "harvester-public"
}

data "harvester_image" "opensuse" {
  namespace    = "harvester-public"
  display_name = "openSUSE-Leap-42.1-OpenStack.x86_64.qcow2"
}