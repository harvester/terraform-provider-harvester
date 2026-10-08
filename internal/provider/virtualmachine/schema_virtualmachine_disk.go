package virtualmachine

import (
	"strings"

	"github.com/harvester/harvester/pkg/builder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// suppressForEmptyCDRom ignores a change on a field of the disk volume when the
// disk is an empty cd-rom: it has no volume, so the field is read back empty
// whatever the configuration says.
func suppressForEmptyCDRom(k, _, _ string, d *schema.ResourceData) bool {
	disk := k[:strings.LastIndex(k, ".")+1]
	return d.Get(disk+constants.FieldDiskType) == builder.DiskTypeCDRom &&
		d.Get(disk+constants.FieldVolumeImage) == "" &&
		d.Get(disk+constants.FieldDiskExistingVolumeName) == "" &&
		d.Get(disk+constants.FieldDiskContainerImageName) == ""
}

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
		constants.FieldDiskCacheMode: {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
			ValidateFunc: validation.StringInSlice([]string{
				constants.DiskCacheModeNone,
				constants.DiskCacheModeWriteBack,
				constants.DiskCacheModeWriteThrough,
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
	}
	volumeFields := []string{
		constants.FieldDiskSize,
		constants.FieldDiskAutoDelete,
		constants.FieldDiskHotPlug,
		constants.FieldVolumeStorageClassName,
		constants.FieldVolumeMode,
		constants.FieldVolumeAccessMode,
	}
	for _, field := range volumeFields {
		s[field].DiffSuppressFunc = suppressForEmptyCDRom
	}
	return s
}
