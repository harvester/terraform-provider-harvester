# VM with a ConfigMap mounted as a disk (application config injection)
resource "harvester_virtualmachine" "with_config" {
  name      = "app-with-config"
  namespace = "default"

  description = "VM with application config injected via ConfigMap disk"

  cpu    = 2
  memory = "4Gi"

  run_strategy = "RerunOnFailure"
  machine_type = "q35"

  network_interface {
    name = "nic-1"
  }

  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "20Gi"
    bus        = "virtio"
    boot_order = 1
    image      = "default/ubuntu-24.04"
  }

  # ConfigMap mounted as a read-only disk inside the VM
  disk {
    name           = "config-disk"
    type           = "disk"
    bus            = "virtio"
    configmap_name = kubernetes_config_map.app_config.metadata[0].name
  }
}

# VM with a Secret mounted as a disk (certificates, credentials)
resource "harvester_virtualmachine" "with_secret_disk" {
  name      = "app-with-certs"
  namespace = "default"

  description = "VM with TLS certificates injected via Secret disk"

  cpu    = 2
  memory = "4Gi"

  run_strategy = "RerunOnFailure"
  machine_type = "q35"

  network_interface {
    name = "nic-1"
  }

  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "20Gi"
    bus        = "virtio"
    boot_order = 1
    image      = "default/ubuntu-24.04"
  }

  # Secret mounted as a read-only disk inside the VM
  disk {
    name        = "certs-disk"
    type        = "disk"
    bus         = "virtio"
    secret_name = "tls-certificates"
  }
}

resource "kubernetes_config_map" "app_config" {
  metadata {
    name      = "app-config"
    namespace = "default"
  }

  data = {
    "app.conf"   = "listen_port=8080\nlog_level=info"
    "db.conf"    = "host=db.internal\nport=5432"
  }
}
