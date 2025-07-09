package importer

import (
	loadbalancerv1 "github.com/harvester/harvester-load-balancer/pkg/apis/loadbalancer.harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceIPPoolStateGetter(obj *loadbalancerv1.IPPool) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName: obj.Name,
	}
	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeIPPool,
		States:       states,
	}, nil
}
