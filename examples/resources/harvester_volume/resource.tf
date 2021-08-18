resource "harvester_volume" "ubuntu20-dev-mount-disk" {
  name = "ubuntu20-dev-mount-disk"
  size = "10Gi"
}

resource "harvester_volume" "ubuntu20-dev-image-disk" {
  name  = "ubuntu20-dev-image-disk"
  size  = "10Gi"
  image = "harvester-public/ubuntu20"
}