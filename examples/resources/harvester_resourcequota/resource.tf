# Harvester ResourceQuota limits VM snapshot sizes within a namespace.
#
# Note: Harvester enforces the resource name to be "default-resource-quota"
# (there is one ResourceQuota per namespace).
resource "harvester_resourcequota" "snapshot_limits" {
  name      = "default-resource-quota"
  namespace = "default"

  # Total snapshot size allowed for the whole namespace, in bytes (here 100 GiB).
  namespace_total_snapshot_size_quota = 107374182400

  # Optional: per-VM snapshot size quota in bytes, keyed by VM name (here 20 GiB).
  vm_total_snapshot_size_quota = {
    "my-vm" = 21474836480
  }
}
