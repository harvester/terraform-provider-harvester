package storageclass

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

const (
	LonghornDriverName = "driver.longhorn.io"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldStorageClassVolumeProvisioner: {
			Type:     schema.TypeString,
			Optional: true,
			Default:  LonghornDriverName,
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
		constants.FieldStorageClassAllowedTopologies: {
			Type:        schema.TypeList,
			Description: "Restrict the node topologies where volumes can be dynamically provisioned.",
			Optional:    true,
			ForceNew:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldStorageClassMatchLabelExpressions: {
						Type:        schema.TypeList,
						Description: "A list of topology selector requirements by labels.",
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"key": {
									Type:        schema.TypeString,
									Description: "The label key that the selector applies to.",
									Optional:    true,
								},
								"values": {
									Type:        schema.TypeSet,
									Description: "An array of string values. One value must match the label to be selected.",
									Optional:    true,
									Elem:        &schema.Schema{Type: schema.TypeString},
								},
							},
						},
					},
				},
			},
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	s := util.DataSourceSchemaWrap(Schema())
	return s
}
