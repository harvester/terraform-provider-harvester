// Package constants defines constants used by the harvester_schedule_backup resource.
package constants

const (
	// ResourceTypeScheduleBackup is the Terraform resource type name for harvester_schedule_backup.
	// This resource manages VM-level backup schedules.
	ResourceTypeScheduleBackup = "harvester_schedule_backup"

	// FieldScheduleBackupVMName is the field name for the VM to backup (required).
	// Format: "namespace/name" or "name" (if in default namespace).
	FieldScheduleBackupVMName = "vm_name"

	// FieldScheduleBackupVolumeName is the deprecated field name for volume-based backup.
	// Deprecated: Use vm_name instead for VM-level backups.
	FieldScheduleBackupVolumeName = "volume_name"

	// FieldScheduleBackupSchedule is the cron schedule for the backup (required, UTC timezone).
	FieldScheduleBackupSchedule = "schedule"

	// FieldScheduleBackupRetain is the number of backups to retain (optional, default: 5).
	FieldScheduleBackupRetain = "retain"

	// FieldScheduleBackupConcurrency is the number of concurrent backup jobs (optional, default: 1).
	// Note: Currently not used by Harvester's ScheduleVMBackup but kept for API compatibility.
	FieldScheduleBackupConcurrency = "concurrency"

	// FieldScheduleBackupLabels are labels to apply to the backup job (optional).
	FieldScheduleBackupLabels = "labels"

	// FieldScheduleBackupTask is the backup task type (currently unused, kept for compatibility).
	FieldScheduleBackupTask = "task"

	// FieldScheduleBackupGroups are groups for the backup job (currently unused, kept for compatibility).
	FieldScheduleBackupGroups = "groups"

	// FieldScheduleBackupCron is an alias for schedule (currently unused, kept for compatibility).
	FieldScheduleBackupCron = "cron"

	// FieldScheduleBackupEnabled controls whether the backup job is enabled (optional, default: true).
	FieldScheduleBackupEnabled = "enabled"
)

const (
	// Longhorn annotation keys for recurring jobs (legacy, not used in current implementation).
	// The current implementation uses Harvester's ScheduleVMBackup CRD instead.
	AnnotationRecurringJobBackup       = "recurring-job.longhorn.io/backup"
	AnnotationRecurringJobBackupRemove = "recurring-job.longhorn.io/backup-remove"
	AnnotationRecurringJobGroupPrefix  = "recurring-job-group.longhorn.io/"

	// NamespaceLonghornSystem is the Kubernetes namespace where Longhorn resources are created.
	NamespaceLonghornSystem = "longhorn-system"

	// Longhorn API constants (legacy, not used in current implementation).
	LonghornAPIGroup             = "longhorn.io"
	LonghornAPIVersion           = "v1beta2"
	LonghornResourceRecurringJob = "recurringjobs"
)
