resource "harvester_virtualmachine" "k3os" {
  count     = 3
  name      = "k3os-${count.index}"
  namespace = "default"

  description = "test k3os iso image"
  tags = {
    ssh-user = "rancher"
  }

  cpu    = 4
  memory = "4Gi"

  efi         = true
  secure_boot = false

  network_interface {
    name         = "nic-1"
    network_name = harvester_network.mgmt-vlan1.id
  }

  disk {
    name       = "cdrom-disk"
    type       = "cd-rom"
    size       = "10Gi"
    bus        = "sata"
    boot_order = 2

    image       = harvester_image.k3os.id
    auto_delete = true
  }

  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "10Gi"
    bus        = "virtio"
    boot_order = 1
  }

  input {
    name = "tablet"
    type = "tablet"
    bus  = "usb"
  }
}


resource "harvester_virtualmachine" "ubuntu20" {
  name                 = "ubuntu20"
  namespace            = "default"
  restart_after_update = true

  description = "test ubuntu20 raw image"
  tags = {
    ssh-user = "ubuntu"
  }

  cpu    = 2
  memory = "2Gi"

  efi         = true
  secure_boot = true

  run_strategy = "RerunOnFailure"
  hostname     = "ubuntu20"
  machine_type = "q35"

  network_interface {
    name           = "nic-1"
    wait_for_lease = true
  }

  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "10Gi"
    bus        = "virtio"
    boot_order = 1

    image       = harvester_image.ubuntu20.id
    auto_delete = true
  }

  disk {
    name        = "emptydisk"
    type        = "disk"
    size        = "20Gi"
    bus         = "virtio"
    auto_delete = true
  }

  cloudinit {
    user_data    = <<-EOF
      #cloud-config
      password: 123456
      chpasswd:
        expire: false
      ssh_pwauth: true
      package_update: true
      packages:
        - qemu-guest-agent
      runcmd:
        - - systemctl
          - enable
          - '--now'
          - qemu-guest-agent
      EOF
    network_data = ""
  }
}

resource "harvester_virtualmachine" "opensuse154" {
  name                 = "opensuse154"
  namespace            = "default"
  restart_after_update = true

  description = "test raw image"
  tags = {
    ssh-user = "opensuse"
  }

  cpu    = 2
  memory = "2Gi"

  efi         = true
  secure_boot = true

  run_strategy = "RerunOnFailure"
  hostname     = "opensuse154"
  machine_type = "q35"

  ssh_keys = [
    harvester_ssh_key.mysshkey.id
  ]

  network_interface {
    name           = "nic-1"
    network_name   = harvester_network.cluster-vlan1.id
    wait_for_lease = true
  }

  network_interface {
    name         = "nic-2"
    model        = "virtio"
    type         = "bridge"
    network_name = harvester_network.cluster-vlan["2"].id
  }

  network_interface {
    name         = "nic-3"
    model        = "e1000"
    type         = "bridge"
    network_name = harvester_network.cluster-vlan["3"].id
  }

  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "10Gi"
    bus        = "virtio"
    boot_order = 1

    image       = harvester_image.opensuse154.id
    auto_delete = true
  }

  disk {
    name = "mount-disk"
    type = "disk"
    bus  = "scsi"

    existing_volume_name = harvester_volume.mount-disk.name
    auto_delete          = false
    hot_plug             = true
  }

  cloudinit {
    user_data    = <<-EOF
      #cloud-config
      password: 123456
      chpasswd:
        expire: false
      ssh_pwauth: true
      package_update: true
      packages:
        - qemu-guest-agent
      runcmd:
        - - systemctl
          - enable
          - '--now'
          - qemu-guest-agent
      ssh_authorized_keys:
        - >-
          public_key content of harvester_ssh_key.mysshkey
      EOF
    network_data = ""
  }
}