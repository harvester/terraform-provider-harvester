package importer

import (
	"strings"

	"github.com/harvester/harvester/pkg/builder"
	"github.com/harvester/harvester/pkg/ref"
	corev1 "k8s.io/api/core/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceVolumeStateGetter(obj *corev1.PersistentVolumeClaim) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonNamespace:   obj.Namespace,
		constants.FieldCommonName:        obj.Name,
		constants.FieldCommonDescription: GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:        GetTags(obj.Labels),
		constants.FieldVolumeSize:        obj.Spec.Resources.Requests.Storage().String(),
		constants.FieldPhase:             obj.Status.Phase,
	}
	if obj.Spec.VolumeMode != nil {
		states[constants.FieldVolumeMode] = obj.Spec.VolumeMode
	}
	if obj.Spec.StorageClassName != nil {
		states[constants.FieldVolumeStorageClassName] = obj.Spec.StorageClassName
	}
	if len(obj.Spec.AccessModes) > 0 {
		states[constants.FieldVolumeAccessMode] = obj.Spec.AccessModes[0]
	}
	if imageID := obj.Annotations[builder.AnnotationKeyImageID]; imageID != "" {
		imageNamespacedName, err := helper.RebuildNamespacedName(imageID, obj.Namespace)
		if err != nil {
			return nil, err
		}
		states[constants.FieldVolumeImage] = imageNamespacedName
	}
	owners, err := ref.GetSchemaOwnersFromAnnotation(obj)
	if err != nil {
		return nil, err
	}
	attachedList := owners.List(kubevirtv1.VirtualMachineGroupVersionKind.GroupKind())
	if len(attachedList) > 0 {
		states[constants.FieldCommonState] = constants.StateVolumeInUse
		states[constants.FieldVolumeAttachedVM] = strings.Join(attachedList, ",")
	} else {
		states[constants.FieldCommonState] = constants.StateCommonReady
	}
	return &StateGetter{
		ID:           helper.BuildID(obj.Namespace, obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeVolume,
		States:       states,
	}, nil
}
