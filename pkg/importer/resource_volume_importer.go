package importer

import (
	"context"
	"strings"

	"github.com/harvester/harvester/pkg/builder"
	"github.com/harvester/harvester/pkg/ref"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceVolumeStateGetter(client *client.Client, obj *corev1.PersistentVolumeClaim) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonNamespace:   obj.Namespace,
		constants.FieldCommonName:        obj.Name,
		constants.FieldCommonDescription: GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:        GetTags(obj.Labels),
		constants.FieldCommonLabels:      GetLabels(obj.Labels),
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

	vms, err := client.HarvesterClient.
		KubevirtV1().
		VirtualMachines(obj.Namespace).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	attachedList := []string{}
	for _, vm := range vms.Items {
		for _, vol := range vm.Spec.Template.Spec.Volumes {
			if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName != "" {
				attachedList = append(attachedList, ref.Construct(obj.Namespace, obj.Name))
			}
		}
	}

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
