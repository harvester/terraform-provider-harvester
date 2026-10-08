package importer

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceKubeOVNVlanStateGetter(obj *kubeovnv1.Vlan) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:               obj.Name,
		constants.FieldCommonDescription:        GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:               GetTags(obj.Labels),
		constants.FieldCommonLabels:             GetLabels(obj.Labels),
		constants.FieldKubeOVNVlanID:            obj.Spec.ID,
		constants.FieldKubeOVNVlanProvider:      obj.Spec.Provider,
		constants.FieldKubeOVNVlanStatusSubnets: obj.Status.Subnets,
	}

	states[constants.FieldCommonState] = constants.StateCommonActive

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKubeOVNVlan,
		States:       states,
	}, nil
}
