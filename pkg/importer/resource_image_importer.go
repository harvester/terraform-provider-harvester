package importer

import (
	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	corev1 "k8s.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/builder"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func ResourceImageStateGetter(obj *harvsterv1.VirtualMachineImage) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonNamespace:   obj.Namespace,
		constants.FieldCommonName:        obj.Name,
		constants.FieldCommonDescription: GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:        GetTags(obj.Labels),
		constants.FieldImageDisplayName:  obj.Spec.DisplayName,
		constants.FieldImageURL:          obj.Spec.URL,
		constants.FieldImageSize:         obj.Status.Size,
	}
	var state string
	for _, condition := range obj.Status.Conditions {
		if condition.Type == harvsterv1.ImageInitialized {
			if condition.Status == corev1.ConditionTrue {
				state = constants.StateCommonActive
			} else if condition.Status == corev1.ConditionFalse {
				state = constants.StateImageFailed
			}
		} else {
			state = constants.StateImageInProgress
		}
	}
	states[constants.FieldCommonState] = state
	return &StateGetter{
		ID:           builder.BuildID(obj.Namespace, obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeImage,
		States:       states,
	}, nil
}
