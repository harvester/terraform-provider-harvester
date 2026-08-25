package importer

import (
	"strings"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	harvsterutil "github.com/harvester/harvester/pkg/util"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

// addonUserLabels drops the addon.harvesterhci.io/* labels (e.g. the
// experimental label set on provider-created addons): they are managed by the
// provider and the Harvester webhook, not by the user labels block, and would
// otherwise drift on every plan.
func addonUserLabels(labels map[string]string) map[string]string {
	userLabels := map[string]string{}
	for key, value := range labels {
		if strings.HasPrefix(key, harvsterutil.AddonPrefix+"/") {
			continue
		}
		userLabels[key] = value
	}
	return userLabels
}

func ResourceAddonStateGetter(obj *harvsterv1.Addon) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonNamespace:    obj.Namespace,
		constants.FieldCommonName:         obj.Name,
		constants.FieldCommonDescription:  GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:         GetTags(obj.Labels),
		constants.FieldCommonLabels:       GetLabels(addonUserLabels(obj.Labels)),
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
