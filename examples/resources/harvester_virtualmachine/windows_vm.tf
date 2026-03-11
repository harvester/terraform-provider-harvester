# Windows Server VM with Hyper-V enlightenments and clock configuration
resource "harvester_virtualmachine" "windows_server" {
  name      = "win-server-2025"
  namespace = "default"

  description = "Windows Server 2025 with Hyper-V optimizations"

  cpu    = 4
  memory = "8Gi"

  run_strategy = "RerunOnFailure"
  machine_type = "q35"
  efi          = true
  secure_boot  = true

  # Hyper-V enlightenments for Windows guests
  hyperv {
    relaxed           = true
    vapic             = true
    vpindex           = true
    runtime           = true
    synic             = true
    reset             = true
    frequencies       = true
    reenlightenment   = true
    tlbflush          = true
    ipi               = true
    spinlocks         = true
    spinlocks_retries = 8192
    synictimer        = true
    synictimer_direct = true
    vendorid          = true
    vendorid_value    = "KVMKVMKVM"
  }

  # Windows clock with local timezone and timers
  clock {
    timezone = "America/New_York"

    timer {
      hpet {
        enabled = false
      }
      pit {
        enabled     = true
        tick_policy = "delay"
      }
      rtc {
        enabled     = true
        tick_policy = "catchup"
        track       = "guest"
      }
      hyperv {
        enabled = true
      }
    }
  }

  network_interface {
    name = "nic-1"
  }

  # OS disk
  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "80Gi"
    bus        = "virtio"
    boot_order = 1
    image      = "default/windows-server-2025"
  }

  # Sysprep answer file from a Kubernetes secret
  disk {
    name                = "sysprep"
    type                = "disk"
    bus                 = "virtio"
    sysprep_secret_name = "win-unattend"
  }
}

# Alternative: Use hyperv_passthrough to enable ALL Hyper-V flags at once
resource "harvester_virtualmachine" "windows_passthrough" {
  name      = "win-passthrough"
  namespace = "default"

  description = "Windows VM with all Hyper-V flags via passthrough"

  cpu    = 4
  memory = "8Gi"

  run_strategy = "RerunOnFailure"
  machine_type = "q35"

  # Mutually exclusive with the hyperv block
  hyperv_passthrough = true

  clock {
    utc_offset_seconds = -18000

    timer {
      hyperv {
        enabled = true
      }
    }
  }

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

# Linux VM with UTC clock and custom timers
resource "harvester_virtualmachine" "linux_custom_clock" {
  name      = "linux-custom-clock"
  namespace = "default"

  description = "Linux VM with explicit clock and timer configuration"

  cpu    = 2
  memory = "4Gi"

  run_strategy = "RerunOnFailure"
  machine_type = "q35"

  clock {
    timezone = "Europe/Paris"

    timer {
      kvm {
        enabled = true
      }
      rtc {
        enabled     = true
        tick_policy = "catchup"
        track       = "wall"
      }
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
