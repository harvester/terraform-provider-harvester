# Look up an image by its Kubernetes name
data "harvester_image" "by_name" {
  name      = "image-nhtf9"
  namespace = "default"
}

# Look up an image by its display name (user-friendly)
# Errors if multiple images share the same display name.
data "harvester_image" "opensuse_tw" {
  namespace    = "default"
  display_name = "openSUSE-Tumbleweed-Minimal-VM.x86_64-Cloud.qcow2"
}

# Use the data source to reference an image in a VM disk
resource "harvester_virtualmachine" "example" {
  name      = "example"
  namespace = "default"

  cpu    = 2
  memory = "4Gi"

  network_interface {
    name = "nic-1"
  }

  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "20Gi"
    bus        = "virtio"
    boot_order = 1

    image       = data.harvester_image.opensuse_tw.id
    auto_delete = true
  }
}
