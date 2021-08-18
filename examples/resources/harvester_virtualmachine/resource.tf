resource "harvester_virtualmachine" "k3os" {
  depends_on = [harvester_image.k3os, harvester_network.vlan1]

  count       = 3
  name        = "k3os-${count.index}"
  description = "test k3os iso"
  tags = {
    ssh-user = "rancher"
  }

  cpu    = 4
  memory = "4Gi"

  network_interface {
    name         = "nic-1"
    network_name = "harvester-public/vlan1"
  }

  disk {
    name       = "cdrom-disk"
    type       = "cd-rom"
    size       = "10Gi"
    bus        = "sata"
    boot_order = 2

    image = "harvester-public/k3os"
  }

  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "10Gi"
    bus        = "virtio"
    boot_order = 1
  }
}


resource "harvester_virtualmachine" "ubuntu20-dev" {
  depends_on = [harvester_image.ubuntu20, harvester_volume.ubuntu20-dev-mount-disk, harvester_network.vlan1, harvester_network.vlan2, harvester_network.vlan3]

  name        = "ubuntu-dev"
  description = "test raw image"

  tags = {
    ssh-user = "ubuntu"
  }

  cpu    = 2
  memory = "2Gi"

  start        = true
  hostname     = "ubuntu-dev"
  machine_type = "q35"

  ssh_keys = [
    "mysshkey"
  ]

  network_interface {
    name         = "nic-1"
    network_name = "harvester-public/vlan1"
  }

  network_interface {
    name         = "nic-2"
    model        = "virtio"
    type         = "bridge"
    network_name = "harvester-public/vlan2"
  }

  network_interface {
    name         = "nic-3"
    model        = "e1000"
    type         = "bridge"
    network_name = "harvester-public/vlan3"
  }

  disk {
    name       = "rootdisk"
    type       = "disk"
    size       = "10Gi"
    bus        = "virtio"
    boot_order = 1

    image = "harvester-public/ubuntu20"
  }

  disk {
    name = "emptydisk"
    type = "disk"
    size = "20Gi"
    bus  = "virtio"
  }

  disk {
    name = "mount-disk"
    type = "disk"
    bus  = "virtio"

    existing_volume_name = "ubuntu20-dev-mount-disk"
    auto_delete          = false
  }

  cloudinit {
    user_data    = <<-EOF
      #cloud-config
      user: ubuntu
      password: root
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
          your ssh public key
      EOF
    network_data = ""
  }
}