# Create a VM template for standardized deployments
resource "harvester_virtualmachinetemplate" "linux_server" {
  name      = "linux-server"
  namespace = "default"

  description = "Standard Linux server template"

  tags = {
    os   = "linux"
    tier = "production"
  }
}

# Create a version of the template with specific VM settings
resource "harvester_virtualmachinetemplateversion" "linux_server_v1" {
  name      = "linux-server-v1"
  namespace = "default"

  description = "Linux server v1 - 2 CPU, 4Gi RAM, openSUSE Tumbleweed"

  template_id = harvester_virtualmachinetemplate.linux_server.id
  image_id    = data.harvester_image.opensuse_tw.id

  cpu    = 2
  memory = "4Gi"

  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "40Gi"
    bus        = "virtio"
    boot_order = 1
  }

  network_interface {
    name = "nic-1"
  }
}

# Create a second version with more resources
resource "harvester_virtualmachinetemplateversion" "linux_server_v2" {
  name      = "linux-server-v2"
  namespace = "default"

  description = "Linux server v2 - 4 CPU, 8Gi RAM"

  template_id = harvester_virtualmachinetemplate.linux_server.id
  image_id    = data.harvester_image.opensuse_tw.id

  cpu    = 4
  memory = "8Gi"

  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "80Gi"
    bus        = "virtio"
    boot_order = 1
  }

  network_interface {
    name = "nic-1"
  }
}

# Look up the image by display name
data "harvester_image" "opensuse_tw" {
  namespace    = "default"
  display_name = "openSUSE-Tumbleweed-Minimal-VM.x86_64-Cloud.qcow2"
}
