package storageclass

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	longhorntypes "github.com/longhorn/longhorn-manager/types"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldStorageClassVolumeProvisioner: {
			Type:     schema.TypeString,
			Optional: true,
			Default:  longhorntypes.LonghornDriverName,
		},
		constants.FieldStorageClassReclaimPolicy: {
			Type:     schema.TypeString,
			Optional: true,
			Default:  string(corev1.PersistentVolumeReclaimDelete),
			ValidateFunc: validation.StringInSlice([]string{
				string(corev1.PersistentVolumeReclaimDelete),
				string(corev1.PersistentVolumeReclaimRetain),
				string(corev1.PersistentVolumeReclaimRecycle),
			}, false),
		},
		constants.FieldStorageClassVolumeBindingMode: {
			Type:     schema.TypeString,
			Optional: true,
			Default:  string(storagev1.VolumeBindingImmediate),
			ValidateFunc: validation.StringInSlice([]string{
				string(storagev1.VolumeBindingImmediate),
				string(storagev1.VolumeBindingWaitForFirstConsumer),
			}, false),
		},
		constants.FieldStorageClassAllowVolumeExpansion: {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		constants.FieldStorageClassIsDefault: {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		constants.FieldStorageClassParameters: {
			Type:        schema.TypeMap,
			Required:    true,
			Description: "refer to https://longhorn.io/docs/latest/volumes-and-nodes/storage-tags. \"migratable\": \"true\" is required for Harvester Virtual Machine LiveMigration",
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	s := util.DataSourceSchemaWrap(Schema())
	return s
}
