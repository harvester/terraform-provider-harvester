package volume

import (
	"github.com/harvester/harvester/pkg/builder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldVolumeSize: {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "1Gi",
		},
		constants.FieldVolumeImage: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldVolumeStorageClassName: {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: util.IsValidName,
		},
		constants.FieldVolumeMode: {
			Type:     schema.TypeString,
			Optional: true,
			Default:  builder.PersistentVolumeModeBlock,
			ValidateFunc: validation.StringInSlice([]string{
				builder.PersistentVolumeModeBlock,
				builder.PersistentVolumeModeFilesystem,
			}, false),
		},
		constants.FieldVolumeAccessMode: {
			Type:     schema.TypeString,
			Optional: true,
			Default:  builder.PersistentVolumeAccessModeReadWriteMany,
			ValidateFunc: validation.StringInSlice([]string{
				builder.PersistentVolumeAccessModeReadWriteOnce,
				builder.PersistentVolumeAccessModeReadOnlyMany,
				builder.PersistentVolumeAccessModeReadWriteMany,
			}, false),
		},
		constants.FieldVolumeAttachedVM: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldPhase: {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
	util.NamespacedSchemaWrap(s, false)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
