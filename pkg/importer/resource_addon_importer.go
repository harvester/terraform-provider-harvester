package importer

import (
	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceAddonStateGetter(obj *harvsterv1.Addon) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonNamespace:    obj.Namespace,
		constants.FieldCommonName:         obj.Name,
		constants.FieldCommonDescription:  GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:         GetTags(obj.Labels),
		constants.FieldCommonLabels:       GetLabels(obj.Labels),
		constants.FieldAddonEnabled:       obj.Spec.Enabled,
		constants.FieldAddonValuesContent: obj.Spec.ValuesContent,
		constants.FieldAddonRepo:          obj.Spec.Repo,
		constants.FieldAddonChart:         obj.Spec.Chart,
		constants.FieldAddonVersion:       obj.Spec.Version,
	}
	states[constants.FieldCommonState] = string(obj.Status.Status)
	states[constants.FieldCommonMessage] = addonMessage(obj)
	return &StateGetter{
		ID:           helper.BuildID(obj.Namespace, obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeAddon,
		States:       states,
	}, nil
}

// addonMessage surfaces the most relevant operation condition message:
// a failure first, then an operation in progress, then the last completion.
func addonMessage(obj *harvsterv1.Addon) string {
	if harvsterv1.AddonOperationFailed.IsTrue(obj) {
		return harvsterv1.AddonOperationFailed.GetMessage(obj)
	}
	if harvsterv1.AddonOperationInProgress.IsTrue(obj) {
		return harvsterv1.AddonOperationInProgress.GetMessage(obj)
	}
	return harvsterv1.AddonOperationCompleted.GetMessage(obj)
}
