# VM with SSH access credentials propagated via cloud-init
resource "harvester_virtualmachine" "ssh_access" {
  name      = "ssh-access-vm"
  namespace = "default"

  description = "VM with SSH key injection via access credentials"

  cpu    = 2
  memory = "4Gi"

  run_strategy = "RerunOnFailure"
  machine_type = "q35"

  # Inject SSH keys from a Kubernetes secret
  access_credentials {
    ssh_public_key {
      secret_name        = kubernetes_secret.ssh_keys.metadata[0].name
      propagation_method = "noCloud"
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

# VM with SSH keys injected via qemu guest agent (runtime update)
resource "harvester_virtualmachine" "guest_agent_ssh" {
  name      = "guest-agent-ssh-vm"
  namespace = "default"

  description = "VM with SSH keys updated at runtime via guest agent"

  cpu    = 2
  memory = "4Gi"

  run_strategy = "RerunOnFailure"
  machine_type = "q35"

  access_credentials {
    ssh_public_key {
      secret_name        = kubernetes_secret.ssh_keys.metadata[0].name
      propagation_method = "qemuGuestAgent"
      users              = ["root", "admin"]
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

# Kubernetes secret containing SSH public keys
resource "kubernetes_secret" "ssh_keys" {
  metadata {
    name      = "my-ssh-keys"
    namespace = "default"
  }

  data = {
    "admin-key" = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5... admin@example.com"
  }
}
