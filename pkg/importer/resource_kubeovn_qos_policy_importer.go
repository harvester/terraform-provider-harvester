package importer

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func flattenBandwidthLimitRules(rules kubeovnv1.QoSPolicyBandwidthLimitRules) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(rules))
	for _, rule := range rules {
		result = append(result, map[string]interface{}{
			constants.FieldKubeOVNQoSRuleName:       rule.Name,
			constants.FieldKubeOVNQoSRuleInterface:  rule.Interface,
			constants.FieldKubeOVNQoSRuleRateMax:    rule.RateMax,
			constants.FieldKubeOVNQoSRuleBurstMax:   rule.BurstMax,
			constants.FieldKubeOVNQoSRulePriority:   rule.Priority,
			constants.FieldKubeOVNQoSRuleDirection:  string(rule.Direction),
			constants.FieldKubeOVNQoSRuleMatchType:  string(rule.MatchType),
			constants.FieldKubeOVNQoSRuleMatchValue: rule.MatchValue,
		})
	}
	return result
}

func ResourceKubeOVNQoSPolicyStateGetter(obj *kubeovnv1.QoSPolicy) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:                    obj.Name,
		constants.FieldCommonDescription:             GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                    GetTags(obj.Labels),
		constants.FieldCommonLabels:                  GetLabels(obj.Labels),
		constants.FieldKubeOVNQoSShared:              obj.Spec.Shared,
		constants.FieldKubeOVNQoSBindingType:         string(obj.Spec.BindingType),
		constants.FieldKubeOVNQoSBandwidthLimitRules: flattenBandwidthLimitRules(obj.Spec.BandwidthLimitRules),
		constants.FieldKubeOVNQoSStatusShared:        obj.Status.Shared,
		constants.FieldKubeOVNQoSStatusBindingType:   string(obj.Status.BindingType),
	}

	states[constants.FieldCommonState] = constants.StateCommonActive

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKubeOVNQoSPolicy,
		States:       states,
	}, nil
}
