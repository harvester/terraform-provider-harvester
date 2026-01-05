// Package constants defines constants used by the harvester_volume_backup resource.
package constants

const (
	// ResourceTypeVolumeBackup is the Terraform resource type name for harvester_volume_backup.
	// Note: Despite the name "volume_backup", this resource manages VM-level backups.
	ResourceTypeVolumeBackup = "harvester_volume_backup"

	// FieldVolumeBackupVMName is the field name for the VM to backup (required).
	// Format: "namespace/name" or "name" (if in default namespace).
	FieldVolumeBackupVMName = "vm_name"
	
	// FieldVolumeBackupVolumeName is the deprecated field name for volume-based backup.
	// Deprecated: Use vm_name instead for VM-level backups.
	FieldVolumeBackupVolumeName = "volume_name"
	
	// FieldVolumeBackupSchedule is the cron schedule for the backup (required, UTC timezone).
	FieldVolumeBackupSchedule = "schedule"
	
	// FieldVolumeBackupRetain is the number of backups to retain (optional, default: 5).
	FieldVolumeBackupRetain = "retain"
	
	// FieldVolumeBackupConcurrency is the number of concurrent backup jobs (optional, default: 1).
	// Note: Currently not used by Harvester's ScheduleVMBackup but kept for API compatibility.
	FieldVolumeBackupConcurrency = "concurrency"
	
	// FieldVolumeBackupLabels are labels to apply to the backup job (optional).
	FieldVolumeBackupLabels = "labels"
	
	// FieldVolumeBackupTask is the backup task type (currently unused, kept for compatibility).
	FieldVolumeBackupTask = "task"
	
	// FieldVolumeBackupGroups are groups for the backup job (currently unused, kept for compatibility).
	FieldVolumeBackupGroups = "groups"
	
	// FieldVolumeBackupCron is an alias for schedule (currently unused, kept for compatibility).
	FieldVolumeBackupCron = "cron"
	
	// FieldVolumeBackupEnabled controls whether the backup job is enabled (optional, default: true).
	FieldVolumeBackupEnabled = "enabled"
)

const (
	// Longhorn annotation keys for recurring jobs (legacy, not used in current implementation).
	// The current implementation uses Harvester's ScheduleVMBackup CRD instead.
	AnnotationRecurringJobBackup      = "recurring-job.longhorn.io/backup"
	AnnotationRecurringJobBackupRemove = "recurring-job.longhorn.io/backup-remove"
	AnnotationRecurringJobGroupPrefix = "recurring-job-group.longhorn.io/"
	
	// NamespaceLonghornSystem is the Kubernetes namespace where Longhorn resources are created.
	NamespaceLonghornSystem = "longhorn-system"
	
	// Longhorn API constants (legacy, not used in current implementation).
	LonghornAPIGroup   = "longhorn.io"
	LonghornAPIVersion = "v1beta2"
	LonghornResourceRecurringJob = "recurringjobs"
)

