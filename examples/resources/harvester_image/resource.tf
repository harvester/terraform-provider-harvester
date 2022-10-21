resource "harvester_image" "k3os" {
  name      = "k3os"
  namespace = "harvester-public"

  display_name = "k3os"
  source_type  = "download"
  url          = "https://github.com/rancher/k3os/releases/download/v0.20.6-k3s1r0/k3os-amd64.iso"
}

resource "harvester_image" "ubuntu20" {
  name      = "ubuntu20"
  namespace = "harvester-public"

  display_name = "ubuntu-20.04-server-cloudimg-amd64.img"
  source_type  = "download"
  url          = "http://cloud-images.ubuntu.com/releases/focal/release/ubuntu-20.04-server-cloudimg-amd64.img"
}

resource "harvester_image" "ubuntu20-any-1" {
  name      = "ubuntu20-any-1"
  namespace = "harvester-public"

  storage_class_name = harvester_storageclass.any-replicas-1.name

  display_name = "ubuntu20-any-1"
  source_type  = "download"
  url          = "http://cloud-images.ubuntu.com/releases/focal/release/ubuntu-20.04-server-cloudimg-amd64.img"
}

resource "harvester_image" "opensuse154" {
  name      = "opensuse154"
  namespace = "harvester-public"

  display_name = "openSUSE-Leap-15.4.x86_64-NoCloud.qcow2"
  source_type  = "download"
  url          = "https://downloadcontent-us1.opensuse.org/repositories/Cloud:/Images:/Leap_15.4/images/openSUSE-Leap-15.4.x86_64-NoCloud.qcow2"
}

resource "harvester_image" "opensuse154-ssd-3" {
  name      = "opensuse154-ssd-3"
  namespace = "harvester-public"

  storage_class_name = harvester_storageclass.ssd-replicas-3.name

  display_name = "opensuse154-ssd-3"
  source_type  = "download"
  url          = "https://downloadcontent-us1.opensuse.org/repositories/Cloud:/Images:/Leap_15.4/images/openSUSE-Leap-15.4.x86_64-NoCloud.qcow2"
}