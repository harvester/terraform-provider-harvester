package volume

import (
	"fmt"

	"github.com/harvester/harvester/pkg/builder"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	Volume *corev1.PersistentVolumeClaim
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().Tags(&c.Volume.Labels).Description(&c.Volume.Annotations)
	customProcessors := []util.Processor{
		{
			Field: constants.FieldVolumeSize,
			Parser: func(i interface{}) error {
				size := i.(string)
				c.Volume.Spec.Resources.Requests[corev1.ResourceStorage] = resource.MustParse(size)
				return nil
			},
		},
		{
			Field: constants.FieldVolumeMode,
			Parser: func(i interface{}) error {
				persistentVolumeMode := corev1.PersistentVolumeBlock
				if volumeMode := i.(string); volumeMode != "" {
					persistentVolumeMode = corev1.PersistentVolumeMode(volumeMode)
				}
				c.Volume.Spec.VolumeMode = &persistentVolumeMode
				return nil
			},
			Required: true,
		},
		{
			Field: constants.FieldVolumeAccessMode,
			Parser: func(i interface{}) error {
				accessModes := []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteMany,
				}
				if accessMode := i.(string); accessMode != "" {
					accessModes = []corev1.PersistentVolumeAccessMode{
						corev1.PersistentVolumeAccessMode(accessMode),
					}
				}
				c.Volume.Spec.AccessModes = accessModes
				return nil
			},
			Required: true,
		},
		{
			Field: constants.FieldVolumeImage,
			Parser: func(i interface{}) error {
				imageNamespacedName := i.(string)
				imageNamespace, imageName, err := helper.NamespacedNamePartsByDefault(imageNamespacedName, c.Volume.Namespace)
				if err != nil {
					return err
				}
				c.Volume.Annotations[builder.AnnotationKeyImageID] = helper.BuildNamespacedName(imageNamespace, imageName)
				storageClassName := builder.BuildImageStorageClassName("", imageName)
				c.Volume.Spec.StorageClassName = ptr.To(storageClassName)
				return nil
			},
		},
		{
			Field: constants.FieldVolumeStorageClassName,
			Parser: func(i interface{}) error {
				storageClassName := i.(string)
				if c.Volume.Annotations[builder.AnnotationKeyImageID] != "" && c.Volume.Spec.StorageClassName != nil && storageClassName != *c.Volume.Spec.StorageClassName {
					return fmt.Errorf("the %s of an image can only be defined during image creation", constants.FieldVolumeStorageClassName)
				} else {
					c.Volume.Spec.StorageClassName = ptr.To(storageClassName)
				}
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
	return c.Volume, nil
}

func newVolumeConstructor(volume *corev1.PersistentVolumeClaim) util.Constructor {
	return &Constructor{
		Volume: volume,
	}
}

func Creator(namespace, name string) util.Constructor {
	volume := &corev1.PersistentVolumeClaim{
		ObjectMeta: util.NewObjectMeta(namespace, name),
		Spec: corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{},
			},
		},
	}
	return newVolumeConstructor(volume)
}

func Updater(volume *corev1.PersistentVolumeClaim) util.Constructor {
	return newVolumeConstructor(volume)
}
