resource "harvester_setting" "backup-target" {
  name = "backup-target"
  value = jsonencode(
    {
      endpoint = "nfs://longhorn-test-nfs-svc.default:/opt/backupstore"
      type     = "nfs"
    }
  )
}

resource "harvester_setting" "default-vm-termination-grace-period-seconds" {
  name  = "default-vm-termination-grace-period-seconds"
  value = "300"
}
