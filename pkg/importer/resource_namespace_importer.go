package importer

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceNamespaceStateGetter(obj *corev1.Namespace) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:        obj.Name,
		constants.FieldCommonDescription: GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:        GetTags(obj.Labels),
		constants.FieldCommonLabels:      GetLabels(obj.Labels),
		constants.FieldCommonState:       string(obj.Status.Phase),
	}
	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeNamespace,
		States:       states,
	}, nil
}
