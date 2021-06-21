package importer

import (
	harvsternetworkv1 "github.com/harvester/harvester-network-controller/pkg/apis/network.harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceClusterNetworkStateGetter(obj *harvsternetworkv1.ClusterNetwork) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonNamespace:                  obj.Namespace,
		constants.FieldCommonName:                       obj.Name,
		constants.FieldCommonDescription:                GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                       GetTags(obj.Labels),
		constants.FieldClusterNetworkEnable:             obj.Enable,
		constants.FieldClusterNetworkDefaultPhysicalNIC: obj.Config[constants.ClusterNetworkConfigKeyDefaultPhysicalNIC],
	}
	return &StateGetter{
		ID:           helper.BuildID(obj.Namespace, obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeClusterNetwork,
		States:       states,
	}, nil
}
