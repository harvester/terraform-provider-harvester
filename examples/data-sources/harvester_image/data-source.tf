data "harvester_image" "ubuntu20" {
  name      = "ubuntu20"
  namespace = "harvester-public"
}

data "harvester_image" "opensuse154" {
  namespace    = "harvester-public"
  display_name = "openSUSE-Leap-15.4.x86_64-NoCloud.qcow2"
}