package volume

import (
	"github.com/harvester/harvester/pkg/builder"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/pointer"
	cdiv1beta1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1beta1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	Volume *cdiv1beta1.DataVolume
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().Tags(&c.Volume.Labels).Description(&c.Volume.Annotations)
	customProcessors := []util.Processor{
		{
			Field: constants.FieldVolumeSize,
			Parser: func(i interface{}) error {
				size := i.(string)
				c.Volume.Spec.PVC.Resources.Requests[corev1.ResourceStorage] = resource.MustParse(size)
				return nil
			},
		},
		{
			Field: constants.FieldVolumeStorageClassName,
			Parser: func(i interface{}) error {
				if storageClassName := i.(string); storageClassName != "" {
					c.Volume.Spec.PVC.StorageClassName = pointer.StringPtr(storageClassName)
				}
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
				c.Volume.Spec.PVC.VolumeMode = &persistentVolumeMode
				return nil
			},
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
				c.Volume.Spec.PVC.AccessModes = accessModes
				return nil
			},
		},
		{
			Field: constants.FieldVolumeImage,
			Parser: func(i interface{}) error {
				if imageNamespacedName := i.(string); imageNamespacedName != "" {
					imageNamespace, imageName, err := helper.NamespacedNamePartsByDefault(imageNamespacedName, c.Volume.Namespace)
					if err != nil {
						return err
					}
					c.Volume.Annotations[builder.AnnotationKeyImageID] = helper.BuildID(imageNamespace, imageName)
					storageClassName := builder.BuildImageStorageClassName(imageNamespace, imageName)
					c.Volume.Spec.PVC.StorageClassName = pointer.StringPtr(storageClassName)
				}
				return nil
			},
		},
	}
	return append(processors, customProcessors...)
}

func (c *Constructor) Result() (interface{}, error) {
	return c.Volume, nil
}

func newVolumeConstructor(volume *cdiv1beta1.DataVolume) util.Constructor {
	return &Constructor{
		Volume: volume,
	}
}

func Creator(namespace, name string) util.Constructor {
	volume := &cdiv1beta1.DataVolume{
		ObjectMeta: util.NewObjectMeta(namespace, name),
		Spec: cdiv1beta1.DataVolumeSpec{
			Source: cdiv1beta1.DataVolumeSource{
				Blank: &cdiv1beta1.DataVolumeBlankImage{},
			},
			PVC: &corev1.PersistentVolumeClaimSpec{
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{},
				},
			},
		},
	}
	return newVolumeConstructor(volume)
}

func Updater(volume *cdiv1beta1.DataVolume) util.Constructor {
	return newVolumeConstructor(volume)
}
