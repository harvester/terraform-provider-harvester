package storageclass

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	StorageClass *storagev1.StorageClass
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().Tags(&c.StorageClass.Labels).Description(&c.StorageClass.Annotations).
		String(constants.FieldStorageClassVolumeProvisioner, &c.StorageClass.Provisioner, true)
	customProcessors := []util.Processor{
		{
			Field: constants.FieldStorageClassReclaimPolicy,
			Parser: func(i interface{}) error {
				reclaimPolicy := corev1.PersistentVolumeReclaimPolicy(i.(string))
				c.StorageClass.ReclaimPolicy = &reclaimPolicy
				return nil
			},
		},
		{
			Field: constants.FieldStorageClassAllowVolumeExpansion,
			Parser: func(i interface{}) error {
				allowVolumeExpansion := i.(bool)
				c.StorageClass.AllowVolumeExpansion = &allowVolumeExpansion
				return nil
			},
		},
		{
			Field: constants.FieldStorageClassVolumeBindingMode,
			Parser: func(i interface{}) error {
				volumeBindingMode := storagev1.VolumeBindingMode(i.(string))
				c.StorageClass.VolumeBindingMode = &volumeBindingMode
				return nil
			},
		},
		{
			Field: constants.FieldStorageClassIsDefault,
			Parser: func(i interface{}) error {
				isDefault := i.(bool)
				isDefaultClass := "false"
				if isDefault {
					isDefaultClass = "true"
				}
				c.StorageClass.Annotations["storageclass.kubernetes.io/is-default-class"] = isDefaultClass
				return nil
			},
		},
		{
			Field: constants.FieldStorageClassParameters,
			Parser: func(i interface{}) error {
				c.StorageClass.Parameters = util.MapMerge(nil, "", i.(map[string]interface{}))
				return nil
			},
		},
		{
			Field: constants.FieldStorageClassAllowedTopologies,
			Parser: func(i interface{}) error {
				if c.StorageClass.AllowedTopologies == nil {
					c.StorageClass.AllowedTopologies = []corev1.TopologySelectorTerm{}
				}
				allowedTopologiesRequirements := []corev1.TopologySelectorLabelRequirement{}
				if itemMap, ok := i.(map[string]interface{}); ok {
					for _, requirement := range itemMap["match_label_expressions"].([]interface{}) {
						if requirementMap, ok := requirement.(map[string]interface{}); ok {
							values := []string{}
							for _, value := range requirementMap["values"].(*schema.Set).List() {
								values = append(values, value.(string))
							}
							allowedTopologiesRequirements = append(allowedTopologiesRequirements, corev1.TopologySelectorLabelRequirement{
								Key:    requirementMap["key"].(string),
								Values: values,
							})
						}
					}
				}
				c.StorageClass.AllowedTopologies = append(c.StorageClass.AllowedTopologies, corev1.TopologySelectorTerm{
					MatchLabelExpressions: allowedTopologiesRequirements,
				})
				return nil
			},
		},
	}
	return append(processors, customProcessors...)
}

func (c *Constructor) Validate() error {
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.StorageClass, nil
}

func newStorageClassConstructor(StorageClass *storagev1.StorageClass) util.Constructor {
	return &Constructor{
		StorageClass: StorageClass,
	}
}

func Creator(name string) util.Constructor {
	storageClass := &storagev1.StorageClass{
		ObjectMeta: util.NewObjectMeta("", name),
	}
	return newStorageClassConstructor(storageClass)
}

func Updater(storageClass *storagev1.StorageClass) util.Constructor {
	return newStorageClassConstructor(storageClass)
}
