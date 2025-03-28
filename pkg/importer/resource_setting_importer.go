package importer

import (
	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	corev1 "k8s.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceSettingStateGetter(obj *harvsterv1.Setting) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:        obj.Name,
		constants.FieldCommonDescription: GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:        GetTags(obj.Labels),
		constants.FieldSettingValue:      obj.Value,
	}

	states[constants.FieldCommonState] = ""
	for _, condition := range obj.Status.Conditions {
		if condition.Type == harvsterv1.SettingConfigured && condition.Status == corev1.ConditionTrue {
			states[constants.FieldCommonState] = constants.StateSettingConfigured
		}
	}
	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeSetting,
		States:       states,
	}, nil
}
