resource "harvester_cloudinit_secret" "cloud-config-opensuse154" {
  name      = "cloud-config-opensuse154"
  namespace = "default"

  depends_on = [
    harvester_ssh_key.mysshkey
  ]

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

resource "harvester_cloudinit_secret" "cloud-config-ubuntu20" {
  name      = "cloud-config-ubuntu20"
  namespace = "default"

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