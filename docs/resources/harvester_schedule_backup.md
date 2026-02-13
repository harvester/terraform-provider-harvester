# harvester_schedule_backup

Manages recurring backup schedules for VirtualMachines in Harvester using ScheduleVMBackup CRDs. This resource creates VM-level backups that include all disks of the VM.

## Example Usage

```hcl
resource "harvester_virtualmachine" "example" {
  name        = "example-vm"
  namespace   = "default"
  cpu         = 2
  memory      = "4Gi"
  run_strategy = "RerunOnFailure"
  hostname     = "example-vm"

  network_interface {
    name         = "nic-1"
    network_name = "default/vlan1"
  }

  disk {
    name       = "disk-1"
    type       = "disk"
    size       = "20Gi"
    bus        = "virtio"
    boot_order = 1
    image      = "harvester-public/image-ubuntu20.04"
  }

  disk {
    name       = "disk-2"
    type       = "disk"
    size       = "10Gi"
    bus        = "virtio"
    boot_order = 2
  }
}

resource "harvester_schedule_backup" "example" {
  name      = "example-vm-backup"
  namespace = "default"

  # The VM to backup (all disks will be included)
  vm_name = "${harvester_virtualmachine.example.namespace}/${harvester_virtualmachine.example.name}"

  # Cron schedule in UTC timezone
  # Format: minute hour day-of-month month day-of-week
  # Example: "0 2 * * *" = Daily at 2 AM UTC
  # IMPORTANT: Harvester uses UTC time, not local time
  schedule = "0 2 * * *"  # Daily at 2 AM UTC

  # Number of backups to retain
  retain = 7

  # Enable or disable the backup schedule
  enabled = true

  labels = {
    environment = "production"
    managed-by  = "terraform"
    purpose     = "daily-backup"
  }

  # Ensure the VM is created before creating the backup schedule
  depends_on = [harvester_virtualmachine.example]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the backup schedule resource.
* `namespace` - (Required) The namespace where the resource will be created.
* `vm_name` - (Required) The name of the VirtualMachine to backup. Format: `namespace/name` or just `name` (defaults to the same namespace). This creates a VM-level backup that includes all disks of the VM.
* `schedule` - (Required) Cron schedule for the backup in UTC timezone. Format: `minute hour day-of-month month day-of-week` (e.g., `0 2 * * *` for daily at 2 AM UTC). **IMPORTANT**: Harvester uses UTC time, not local time. Adjust your schedule accordingly.
* `retain` - (Optional) Number of backups to retain. Older backups will be automatically deleted. Minimum: 1, Default: 5.
* `enabled` - (Optional) Whether the backup schedule is enabled. Default: `true`.
* `labels` - (Optional) Map of labels to apply to the ScheduleVMBackup resource.
* `volume_name` - (Optional, Deprecated) The name of the volume to backup. **DEPRECATED**: Use `vm_name` instead for VM-level backups. This field is kept for backward compatibility only.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

* `id` - The unique identifier for the resource. Format: `namespace/vmname/jobname`.
* `message` - Status message from the ScheduleVMBackup (if available).
* `state` - Current state of the backup schedule (if available).

## Notes

### VM-Level Backups

This resource creates **VM-level backups** that include all disks of the VirtualMachine. This is the recommended approach as it ensures consistency across all disks and is visible in the Harvester UI.

### Schedule Format and Timezone

The `schedule` field uses **UTC timezone**. Harvester does not use local time for cron schedules. Common examples:

* `"0 2 * * *"` - Daily at 2 AM UTC
* `"0 0 * * 0"` - Weekly on Sunday at midnight UTC
* `"0 */6 * * *"` - Every 6 hours
* `"0 0 1 * *"` - Monthly on the 1st at midnight UTC

**Important**: If your local time is UTC+1 (CET/CEST), a schedule of `"0 2 * * *"` will run at 3 AM local time (2 AM UTC + 1 hour).

### One Schedule Per VM

Harvester allows only **one ScheduleVMBackup resource per VM**. If you try to create a second schedule for the same VM, Terraform will update the existing schedule instead of creating a new one.

### Schedule Granularity

Harvester enforces a minimum granularity of **1 hour** for backup schedules. Schedules with the same granularity must be at least 10 minutes apart. For example:
- ✅ Valid: `"0 2 * * *"` (daily at 2 AM)
- ✅ Valid: `"0 2 * * *"` and `"10 2 * * *"` (10 minutes apart)
- ❌ Invalid: `"0 2 * * *"` and `"5 2 * * *"` (less than 10 minutes apart)

### Deprecated volume_name Field

The `volume_name` field is deprecated and kept only for backward compatibility. When using `volume_name`, the provider will:
1. Find the VM that uses the specified volume
2. Create a VM-level backup schedule for that VM

It is recommended to use `vm_name` directly instead.

### Backup Retention

The `retain` field controls how many backups are kept. When the limit is reached, older backups are automatically deleted. The minimum value is 1.

### Backup Visibility

Backups created by this resource are visible in:
- Harvester UI (Virtual Machines → Backups)
- Longhorn UI (Backups section)

The ScheduleVMBackup resource ensures proper integration with Harvester's backup management system.

## Import

Backup schedule resources can be imported using the resource ID:

```bash
terraform import harvester_schedule_backup.example default/vm-name/job-name
```

The ID format is: `namespace/vmname/jobname`

## Related Documentation

- [Harvester Backup Documentation](https://docs.harvesterhci.io/v1.7/backup-restore/backup)
- [Longhorn Backup Documentation](https://longhorn.io/docs/1.5.0/snapshots-and-backups/backup-and-restore/)

