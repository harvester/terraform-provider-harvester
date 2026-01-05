// Package volumebackup provides the Terraform schema definitions for the harvester_volume_backup resource.
package volumebackup

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// Schema returns the Terraform schema for the harvester_volume_backup resource.
// This resource manages recurring VM backups using Harvester's ScheduleVMBackup CRD.
// Note: Despite the name "volume_backup", this resource manages VM-level backups (all disks).
func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldVolumeBackupVMName: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the virtual machine to backup. Format: 'namespace/name' or 'name' (if in default namespace). This creates a VM-level backup that includes all disks.",
		},
		constants.FieldVolumeBackupVolumeName: {
			Type:        schema.TypeString,
			Optional:    true,
			Deprecated:  "Use vm_name instead. This field is kept for backward compatibility only.",
			Description: "[DEPRECATED] The name of the volume to backup. Use vm_name for VM-level backups instead.",
		},
		constants.FieldVolumeBackupSchedule: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Cron schedule for the backup in UTC timezone (e.g., '0 0 * * *' for daily at midnight UTC). Harvester uses UTC time, not local time.",
		},
		constants.FieldVolumeBackupRetain: {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     5,
			Description: "Number of backups to retain. Older backups will be automatically deleted. Minimum: 1, Default: 5",
			ValidateFunc: validation.IntAtLeast(1),
		},
		constants.FieldVolumeBackupConcurrency: {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     1,
			Description: "Number of concurrent backup jobs. Note: This field is currently not used by Harvester's ScheduleVMBackup but is kept for API compatibility. Default: 1",
			ValidateFunc: validation.IntBetween(1, 10),
		},
		constants.FieldVolumeBackupLabels: {
			Type:        schema.TypeMap,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Labels to apply to the backup job",
		},
		constants.FieldVolumeBackupGroups: {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Groups for the backup job",
		},
		constants.FieldVolumeBackupEnabled: {
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

