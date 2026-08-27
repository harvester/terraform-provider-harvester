# VM with runtime options for production workloads
resource "harvester_virtualmachine" "production" {
  name      = "production-app"
  namespace = "default"

  description = "Production VM with eviction strategy and graceful shutdown"

  cpu    = 4
  memory = "8Gi"

  run_strategy = "RerunOnFailure"
  machine_type = "q35"

  # Live-migrate on node drain or resource pressure
  eviction_strategy = "LiveMigrate"

  # Allow 5 minutes for graceful shutdown
  termination_grace_period_seconds = 300

  # OS type hint for KubeVirt optimizations
  os_type = "linux"

  network_interface {
    name = "nic-1"
  }

  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "40Gi"
    bus        = "virtio"
    boot_order = 1
    image      = "default/ubuntu-24.04"
  }
}

# Windows VM with appropriate runtime settings
resource "harvester_virtualmachine" "windows_runtime" {
  name      = "windows-server"
  namespace = "default"

  description = "Windows Server with extended shutdown grace period"

  cpu    = 4
  memory = "8Gi"

  run_strategy = "RerunOnFailure"
  machine_type = "q35"

  eviction_strategy = "LiveMigrate"

  # Windows needs more time for graceful shutdown
  termination_grace_period_seconds = 600

  os_type = "windows"

  network_interface {
    name = "nic-1"
  }

  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "80Gi"
    bus        = "virtio"
    boot_order = 1
    image      = "default/windows-server-2025"
  }
}
