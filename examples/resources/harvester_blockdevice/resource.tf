# Adopt and provision a block device for Longhorn storage
resource "harvester_blockdevice" "nvme_data" {
  name      = "blockdevice-pci-0000-04-00-0-abcdef123456"
  namespace = "longhorn-system"

  description = "NVMe data drive on node-01"

  tags = {
    role = "data"
  }

  labels = {
    tier = "fast"
  }

  # Provisioned = device is formatted and used by Longhorn
  provisioned = true

  # Force formatting even if the device has an existing filesystem
  # force_formatted = true
}

# Adopt a device without provisioning (monitoring only)
resource "harvester_blockdevice" "spare_disk" {
  name      = "blockdevice-pci-0000-05-00-0-fedcba654321"
  namespace = "longhorn-system"

  description = "Spare disk - not yet provisioned"

  provisioned = false
}
