package importer

import (
	harvsternetworkv1 "github.com/harvester/harvester-network-controller/pkg/apis/network.harvesterhci.io/v1beta1"
	corev1 "k8s.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceClusterNetworkStateGetter(obj *harvsternetworkv1.ClusterNetwork) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:        obj.Name,
		constants.FieldCommonDescription: GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:        GetTags(obj.Labels),
	}
	states[constants.FieldCommonState] = constants.StateCommonActive
	for _, condition := range obj.Status.Conditions {
		if condition.Type == harvsternetworkv1.Ready && condition.Status == corev1.ConditionTrue {
			states[constants.FieldCommonState] = constants.StateCommonReady
		}
	}
	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeClusterNetwork,
		States:       states,
	}, nil
}
