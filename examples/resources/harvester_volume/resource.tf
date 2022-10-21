resource "harvester_volume" "mount-disk" {
  name      = "mount-disk"
  namespace = "default"

  size = "10Gi"
}

resource "harvester_volume" "mount-ssd-3-disk" {
  name      = "mount-ssd-3-disk"
  namespace = "default"

  storage_class_name = harvester_storageclass.ssd-replicas-3.name

  size = "10Gi"
}

resource "harvester_volume" "mount-any-1-disk" {
  name      = "mount-any-1-disk"
  namespace = "default"

  storage_class_name = harvester_storageclass.any-replicas-1.name

  size = "10Gi"
}

resource "harvester_volume" "ubuntu20-image-disk" {
  name      = "ubuntu20-image-disk"
  namespace = "default"

  size  = "10Gi"
  image = harvester_image.ubuntu20.id
}

resource "harvester_volume" "opensuse154-image-disk" {
  name      = "opensuse154-image-disk"
  namespace = "default"

  size  = "10Gi"
  image = harvester_image.opensuse154.id
}