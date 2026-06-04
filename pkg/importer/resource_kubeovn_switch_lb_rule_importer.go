package importer

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func flattenSlrPorts(ports []kubeovnv1.SwitchLBRulePort) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(ports))
	for _, port := range ports {
		result = append(result, map[string]interface{}{
			constants.FieldKubeOVNSlrPortName:       port.Name,
			constants.FieldKubeOVNSlrPortPort:       int(port.Port),
			constants.FieldKubeOVNSlrPortTargetPort: int(port.TargetPort),
			constants.FieldKubeOVNSlrPortProtocol:   port.Protocol,
		})
	}
	return result
}

func ResourceKubeOVNSwitchLBRuleStateGetter(obj *kubeovnv1.SwitchLBRule) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:                         obj.Name,
		constants.FieldCommonDescription:                  GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                         GetTags(obj.Labels),
		constants.FieldCommonLabels:                       GetLabels(obj.Labels),
		constants.FieldKubeOVNSwitchLBRuleVip:             obj.Spec.Vip,
		constants.FieldKubeOVNSwitchLBRuleNamespace:       obj.Spec.Namespace,
		constants.FieldKubeOVNSwitchLBRuleSelector:        obj.Spec.Selector,
		constants.FieldKubeOVNSwitchLBRuleEndpoints:       obj.Spec.Endpoints,
		constants.FieldKubeOVNSwitchLBRuleSessionAffinity: obj.Spec.SessionAffinity,
		constants.FieldKubeOVNSwitchLBRulePorts:           flattenSlrPorts(obj.Spec.Ports),
		constants.FieldKubeOVNSwitchLBRuleStatusPorts:     obj.Status.Ports,
		constants.FieldKubeOVNSwitchLBRuleStatusService:   obj.Status.Service,
	}

	if len(obj.Status.Service) > 0 {
		states[constants.FieldCommonState] = constants.StateCommonReady
	} else {
		states[constants.FieldCommonState] = constants.StateCommonActive
	}

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKubeOVNSwitchLBRule,
		States:       states,
	}, nil
}
