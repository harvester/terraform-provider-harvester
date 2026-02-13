package schedulebackup

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldScheduleBackupVMName: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the virtual machine to backup. Format: 'namespace/name' or 'name' (if in default namespace).",
		},
		constants.FieldScheduleBackupVolumeName: {
			Type:       schema.TypeString,
			Optional:   true,
			Deprecated: "Use vm_name instead.",
		},
		constants.FieldScheduleBackupSchedule: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Cron schedule for the backup in UTC timezone (e.g., '0 0 * * *' for daily at midnight UTC).",
		},
		constants.FieldScheduleBackupRetain: {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      5,
			Description:  "Number of backups to retain. Minimum: 1, Default: 5",
			ValidateFunc: validation.IntAtLeast(1),
		},
		constants.FieldScheduleBackupLabels: {
			Type:        schema.TypeMap,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Labels to apply to the ScheduleVMBackup resource",
		},
		constants.FieldScheduleBackupEnabled: {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Whether the backup schedule is enabled (default: true). When false, the schedule is suspended.",
		},
	}
	util.NamespacedSchemaWrap(s, false)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
