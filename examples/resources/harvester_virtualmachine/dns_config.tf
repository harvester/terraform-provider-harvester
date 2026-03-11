# VM with custom DNS configuration
resource "harvester_virtualmachine" "custom_dns" {
  name      = "custom-dns-vm"
  namespace = "default"

  description = "VM with custom DNS policy and configuration"

  cpu    = 2
  memory = "4Gi"

  run_strategy = "RerunOnFailure"
  machine_type = "q35"

  # Override cluster DNS with custom nameservers
  dns_policy = "None"

  dns_config {
    nameservers = ["8.8.8.8", "8.8.4.4"]
    searches    = ["example.com", "internal.local"]

    options {
      name  = "ndots"
      value = "5"
    }
    options {
      name = "single-request-reopen"
    }
  }

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
}
