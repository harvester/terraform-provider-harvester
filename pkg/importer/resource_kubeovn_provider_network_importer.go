package importer

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceKubeOVNProviderNetworkStateGetter(obj *kubeovnv1.ProviderNetwork) (*StateGetter, error) {
	customInterfaces := make([]map[string]interface{}, 0, len(obj.Spec.CustomInterfaces))
	for _, ci := range obj.Spec.CustomInterfaces {
		customInterfaces = append(customInterfaces, map[string]interface{}{
			constants.FieldKubeOVNCustomInterfaceInterface: ci.Interface,
			constants.FieldKubeOVNCustomInterfaceNodes:     ci.Nodes,
		})
	}

	states := map[string]interface{}{
		constants.FieldCommonName:                            obj.Name,
		constants.FieldCommonDescription:                     GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                            GetTags(obj.Labels),
		constants.FieldCommonLabels:                          GetLabels(obj.Labels),
		constants.FieldKubeOVNProviderNetDefaultInterface:    obj.Spec.DefaultInterface,
		constants.FieldKubeOVNProviderNetCustomInterfaces:    customInterfaces,
		constants.FieldKubeOVNProviderNetExcludeNodes:        obj.Spec.ExcludeNodes,
		constants.FieldKubeOVNProviderNetExchangeLinkName:    obj.Spec.ExchangeLinkName,
		constants.FieldKubeOVNProviderNetStatusReady:         obj.Status.Ready,
		constants.FieldKubeOVNProviderNetStatusReadyNodes:    obj.Status.ReadyNodes,
		constants.FieldKubeOVNProviderNetStatusNotReadyNodes: obj.Status.NotReadyNodes,
		constants.FieldKubeOVNProviderNetStatusVlans:         obj.Status.Vlans,
	}

	states[constants.FieldCommonState] = constants.StateCommonActive

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKubeOVNProviderNetwork,
		States:       states,
	}, nil
}
