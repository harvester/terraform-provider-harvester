package virtualmachine

import (
	"github.com/harvester/harvester/pkg/builder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func resourceDiskSchema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldDiskName: {
			Type:     schema.TypeString,
			Required: true,
		},
		constants.FieldDiskType: {
			Type:     schema.TypeString,
			Optional: true,
			Default:  builder.DiskTypeDisk,
			ValidateFunc: validation.StringInSlice([]string{
				builder.DiskTypeDisk,
				builder.DiskTypeCDRom,
			}, false),
		},
		constants.FieldDiskSize: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldDiskBus: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ValidateFunc: validation.StringInSlice([]string{
				builder.DiskBusVirtio,
				builder.DiskBusSata,
				builder.DiskBusScsi,
				"",
			}, false),
		},
		constants.FieldDiskBootOrder: {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      0,
			ValidateFunc: validation.IntAtLeast(0),
		},
		constants.FieldVolumeImage: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldDiskExistingVolumeName: {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: util.IsValidName,
		},
		constants.FieldDiskContainerImageName: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldDiskAutoDelete: {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		constants.FieldDiskHotPlug: {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
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
			Computed: true,
			ValidateFunc: validation.StringInSlice([]string{
				builder.PersistentVolumeModeBlock,
				builder.PersistentVolumeModeFilesystem,
			}, false),
		},
		constants.FieldVolumeAccessMode: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ValidateFunc: validation.StringInSlice([]string{
				builder.PersistentVolumeAccessModeReadWriteOnce,
				builder.PersistentVolumeAccessModeReadOnlyMany,
				builder.PersistentVolumeAccessModeReadWriteMany,
			}, false),
		},
		constants.FieldDiskVolumeName: {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: util.IsValidName,
		},
		constants.FieldDiskEject: {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Eject the CD-ROM disk by opening the tray. Only applies to cd-rom type disks.",
		},
	}
	return s
}
