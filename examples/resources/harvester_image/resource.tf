resource "harvester_image" "k3os" {
  name      = "k3os"
  namespace = "harvester-public"

  display_name = "k3os"
  url          = "https://github.com/rancher/k3os/releases/download/v0.20.6-k3s1r0/k3os-amd64.iso"
}

resource "harvester_image" "ubuntu20" {
  name      = "ubuntu20"
  namespace = "harvester-public"

  display_name = "ubuntu20"
  url          = "http://cloud-images.ubuntu.com/releases/focal/release/ubuntu-20.04-server-cloudimg-amd64.img"
}