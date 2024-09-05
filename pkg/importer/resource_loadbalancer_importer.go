package importer

import (
	loadbalancerv1 "github.com/harvester/harvester-load-balancer/pkg/apis/loadbalancer.harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceLoadBalancerStateGetter(obj *loadbalancerv1.LoadBalancer) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonNamespace: obj.Namespace,
		constants.FieldCommonName:      obj.Name,
	}
	return &StateGetter{
		ID:           helper.BuildID(obj.Namespace, obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeLoadBalancer,
		States:       states,
	}, nil
}
