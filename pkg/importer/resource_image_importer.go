package importer

import (
	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	harvsterutil "github.com/harvester/harvester/pkg/util"
	corev1 "k8s.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceImageStateGetter(obj *harvsterv1.VirtualMachineImage) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonNamespace:             obj.Namespace,
		constants.FieldCommonName:                  obj.Name,
		constants.FieldCommonDescription:           GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                  GetTags(obj.Labels),
		constants.FieldImageDisplayName:            obj.Spec.DisplayName,
		constants.FieldImageSourceType:             obj.Spec.SourceType,
		constants.FieldImageURL:                    obj.Spec.URL,
		constants.FieldImagePVCNamespace:           obj.Spec.PVCNamespace,
		constants.FieldImagePVCName:                obj.Spec.PVCName,
		constants.FieldImageProgress:               obj.Status.Progress,
		constants.FieldImageSize:                   obj.Status.Size,
		constants.FieldImageVolumeStorageClassName: obj.Status.StorageClassName,
		constants.FieldImageStorageClassParameters: obj.Spec.StorageClassParameters,
		constants.FieldImageStorageClassName:       obj.Annotations[harvsterutil.AnnotationStorageClassName],
	}

	var (
		state       string
		InitMessage string
		initialized bool
		imported    bool
	)
	for _, condition := range obj.Status.Conditions {
		switch condition.Type {
		case harvsterv1.ImageInitialized:
			initialized = condition.Status == corev1.ConditionTrue
			InitMessage = condition.Message
		case harvsterv1.ImageImported:
			imported = condition.Status == corev1.ConditionTrue
		}
	}
	if initialized {
		if imported {
			state = constants.StateCommonActive
		} else {
			switch obj.Spec.SourceType {
			case harvsterv1.VirtualMachineImageSourceTypeDownload:
				state = constants.StateImageDownloading
			case harvsterv1.VirtualMachineImageSourceTypeExportVolume:
				state = constants.StateImageExporting
			default:
				state = constants.StateImageUploading
			}
		}
	} else if InitMessage == "" {
		state = constants.StateImageInitializing
	} else {
		state = constants.StateCommonFailed
	}
	states[constants.FieldCommonState] = state
	states[constants.FieldCommonMessage] = InitMessage
	return &StateGetter{
		ID:           helper.BuildID(obj.Namespace, obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeImage,
		States:       states,
	}, nil
}
