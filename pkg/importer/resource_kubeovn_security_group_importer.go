package importer

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func flattenSgRules(rules []*kubeovnv1.SgRule) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(rules))
	for _, rule := range rules {
		result = append(result, map[string]interface{}{
			constants.FieldKubeOVNSGRuleIPVersion:           rule.IPVersion,
			constants.FieldKubeOVNSGRuleProtocol:            string(rule.Protocol),
			constants.FieldKubeOVNSGRulePriority:            rule.Priority,
			constants.FieldKubeOVNSGRuleRemoteType:          string(rule.RemoteType),
			constants.FieldKubeOVNSGRuleRemoteAddress:       rule.RemoteAddress,
			constants.FieldKubeOVNSGRuleRemoteSecurityGroup: rule.RemoteSecurityGroup,
			constants.FieldKubeOVNSGRulePortRangeMin:        rule.PortRangeMin,
			constants.FieldKubeOVNSGRulePortRangeMax:        rule.PortRangeMax,
			constants.FieldKubeOVNSGRulePolicy:              string(rule.Policy),
		})
	}
	return result
}

func ResourceKubeOVNSecurityGroupStateGetter(obj *kubeovnv1.SecurityGroup) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:                     obj.Name,
		constants.FieldCommonDescription:              GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                     GetTags(obj.Labels),
		constants.FieldCommonLabels:                   GetLabels(obj.Labels),
		constants.FieldKubeOVNSGAllowSameGroupTraffic: obj.Spec.AllowSameGroupTraffic,
		constants.FieldKubeOVNSGIngressRules:          flattenSgRules(obj.Spec.IngressRules),
		constants.FieldKubeOVNSGEgressRules:           flattenSgRules(obj.Spec.EgressRules),
		constants.FieldKubeOVNSGStatusPortGroup:       obj.Status.PortGroup,
		constants.FieldKubeOVNSGStatusIngressMD5:      obj.Status.IngressMd5,
		constants.FieldKubeOVNSGStatusEgressMD5:       obj.Status.EgressMd5,
		constants.FieldKubeOVNSGStatusIngressSynced:   obj.Status.IngressLastSyncSuccess,
		constants.FieldKubeOVNSGStatusEgressSynced:    obj.Status.EgressLastSyncSuccess,
	}

	states[constants.FieldCommonState] = constants.StateCommonActive

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKubeOVNSecurityGroup,
		States:       states,
	}, nil
}
