// Package schedulebackup provides the Terraform schema definitions for the harvester_schedule_backup resource.
package schedulebackup

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// Schema returns the Terraform schema for the harvester_schedule_backup resource.
// This resource manages recurring VM backups using Harvester's ScheduleVMBackup CRD.
// This resource manages VM-level backup schedules (all disks).
func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldScheduleBackupVMName: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the virtual machine to backup. Format: 'namespace/name' or 'name' (if in default namespace). This creates a VM-level backup that includes all disks.",
		},
		constants.FieldScheduleBackupVolumeName: {
			Type:        schema.TypeString,
			Optional:    true,
			Deprecated:  "Use vm_name instead. This field is kept for backward compatibility only.",
			Description: "[DEPRECATED] The name of the volume to backup. Use vm_name for VM-level backups instead.",
		},
		constants.FieldScheduleBackupSchedule: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Cron schedule for the backup in UTC timezone (e.g., '0 0 * * *' for daily at midnight UTC). Harvester uses UTC time, not local time.",
		},
		constants.FieldScheduleBackupRetain: {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      5,
			Description:  "Number of backups to retain. Older backups will be automatically deleted. Minimum: 1, Default: 5",
			ValidateFunc: validation.IntAtLeast(1),
		},
		constants.FieldScheduleBackupConcurrency: {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      1,
			Description:  "Number of concurrent backup jobs. Note: This field is currently not used by Harvester's ScheduleVMBackup but is kept for API compatibility. Default: 1",
			ValidateFunc: validation.IntBetween(1, 10),
		},
		constants.FieldScheduleBackupLabels: {
			Type:        schema.TypeMap,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Labels to apply to the backup job",
		},
		constants.FieldScheduleBackupGroups: {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Groups for the backup job",
		},
		constants.FieldScheduleBackupEnabled: {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Whether the backup job is enabled (default: true)",
		},
	}
	util.NamespacedSchemaWrap(s, false)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
