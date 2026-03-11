data "harvester_blockdevice" "nvme" {
  name      = "blockdevice-pci-0000-04-00-0-abcdef123456"
  namespace = "longhorn-system"
}
