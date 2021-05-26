package importer

import (
	"strings"

	"github.com/harvester/harvester/pkg/ref"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1beta1"

	"github.com/harvester/terraform-provider-harvester/pkg/builder"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func ResourceVolumeStateGetter(obj *cdiv1beta1.DataVolume) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonNamespace:   obj.Namespace,
		constants.FieldCommonName:        obj.Name,
		constants.FieldCommonDescription: GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:        GetTags(obj.Labels),
		constants.FieldVolumeSize:        obj.Spec.PVC.Resources.Requests.Storage().String(),
		constants.FieldPhase:             obj.Status.Phase,
		constants.FieldProgress:          obj.Status.Progress,
	}
	if obj.Spec.PVC != nil {
		if obj.Spec.PVC.VolumeMode != nil {
			states[constants.FieldVolumeMode] = obj.Spec.PVC.VolumeMode
		}
		if obj.Spec.PVC.StorageClassName != nil {
			states[constants.FieldVolumeStorageClassName] = obj.Spec.PVC.StorageClassName
		}
		if len(obj.Spec.PVC.AccessModes) > 0 {
			states[constants.FieldVolumeAccessMode] = obj.Spec.PVC.AccessModes[0]
		}
	}
	if imageID := obj.Annotations[builder.AnnotationKeyImageID]; imageID != "" {
		imageNamespacedName, err := builder.BuildNamespacedNameFromID(imageID, obj.Namespace)
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
		ID:           builder.BuildID(obj.Namespace, obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeVolume,
		States:       states,
	}, nil
}
