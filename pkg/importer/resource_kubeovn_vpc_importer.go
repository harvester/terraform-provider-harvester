package importer

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceKubeOVNVpcStateGetter(obj *kubeovnv1.Vpc) (*StateGetter, error) {
	staticRoutes := make([]map[string]interface{}, 0, len(obj.Spec.StaticRoutes))
	for _, route := range obj.Spec.StaticRoutes {
		staticRoutes = append(staticRoutes, map[string]interface{}{
			constants.FieldKubeOVNStaticRoutePolicy:    string(route.Policy),
			constants.FieldKubeOVNStaticRouteCIDR:      route.CIDR,
			constants.FieldKubeOVNStaticRouteNextHopIP: route.NextHopIP,
			constants.FieldKubeOVNStaticRouteECMPMode:  route.ECMPMode,
			constants.FieldKubeOVNStaticRouteTable:     route.RouteTable,
		})
	}

	policyRoutes := make([]map[string]interface{}, 0, len(obj.Spec.PolicyRoutes))
	for _, route := range obj.Spec.PolicyRoutes {
		policyRoutes = append(policyRoutes, map[string]interface{}{
			constants.FieldKubeOVNPolicyRoutePriority:  route.Priority,
			constants.FieldKubeOVNPolicyRouteMatch:     route.Match,
			constants.FieldKubeOVNPolicyRouteAction:    string(route.Action),
			constants.FieldKubeOVNPolicyRouteNextHopIP: route.NextHopIP,
		})
	}

	states := map[string]interface{}{
		constants.FieldCommonName:               obj.Name,
		constants.FieldCommonDescription:        GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:               GetTags(obj.Labels),
		constants.FieldCommonLabels:             GetLabels(obj.Labels),
		constants.FieldKubeOVNVpcNamespaces:     obj.Spec.Namespaces,
		constants.FieldKubeOVNVpcStaticRoutes:   staticRoutes,
		constants.FieldKubeOVNVpcPolicyRoutes:   policyRoutes,
		constants.FieldKubeOVNVpcEnableExternal: obj.Spec.EnableExternal,
		constants.FieldKubeOVNVpcEnableBfd:      obj.Spec.EnableBfd,
		constants.FieldKubeOVNVpcDefaultSubnet:  obj.Status.DefaultLogicalSwitch,
		constants.FieldKubeOVNVpcStandby:        obj.Status.Standby,
		constants.FieldKubeOVNVpcRouter:         obj.Status.Router,
		constants.FieldKubeOVNVpcSubnets:        obj.Status.Subnets,
	}

	states[constants.FieldCommonState] = constants.StateCommonActive
	for _, condition := range obj.Status.Conditions {
		if kubeovnv1.ConditionType(condition.Type) == "Ready" && condition.Status == corev1.ConditionStatus("True") {
			states[constants.FieldCommonState] = constants.StateCommonReady
		}
	}

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKubeOVNVpc,
		States:       states,
	}, nil
}
