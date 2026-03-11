# VM with a CD-ROM drive that can be ejected (install media workflow)
resource "harvester_virtualmachine" "install_from_iso" {
  name      = "install-vm"
  namespace = "default"

  description = "VM with ejectable CD-ROM for OS installation"

  cpu    = 2
  memory = "4Gi"

  run_strategy = "RerunOnFailure"
  machine_type = "q35"

  network_interface {
    name = "nic-1"
  }

  # Root disk for the OS installation target
  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "40Gi"
    bus        = "virtio"
    boot_order = 2
  }

  # CD-ROM with ISO image - eject after installation
  disk {
    name       = "install-cd"
    type       = "cd-rom"
    bus        = "sata"
    boot_order = 1

    image = harvester_image.install_iso.id

    # Set to "open" to eject the disc after installation
    # Set to "closed" (default) to keep the disc inserted
    eject = "closed"
  }
}

resource "harvester_image" "install_iso" {
  name         = "tinycore-iso"
  namespace    = "default"
  display_name = "TinyCore Linux ISO"
  source_type  = "download"
  url          = "https://distro.ibiblio.org/tinycorelinux/16.x/x86/release/TinyCore-current.iso"
}
