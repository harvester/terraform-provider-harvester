package importer

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceKubeOVNVpcNatGatewayStateGetter(obj *kubeovnv1.VpcNatGateway) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:                     obj.Name,
		constants.FieldCommonDescription:              GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                     GetTags(obj.Labels),
		constants.FieldCommonLabels:                   GetLabels(obj.Labels),
		constants.FieldKubeOVNVpcNatGwVpc:             obj.Spec.Vpc,
		constants.FieldKubeOVNVpcNatGwSubnet:          obj.Spec.Subnet,
		constants.FieldKubeOVNVpcNatGwLanIP:           obj.Spec.LanIP,
		constants.FieldKubeOVNVpcNatGwExternalSubnets: obj.Spec.ExternalSubnets,
		constants.FieldKubeOVNVpcNatGwSelector:        obj.Spec.Selector,
		constants.FieldKubeOVNVpcNatGwQoSPolicy:       obj.Spec.QoSPolicy,
		constants.FieldKubeOVNVpcNatGwStatusQoS:       obj.Status.QoSPolicy,
		constants.FieldKubeOVNVpcNatGwStatusExtSubs:   obj.Status.ExternalSubnets,
		constants.FieldKubeOVNVpcNatGwStatusSelector:  obj.Status.Selector,
	}

	// VpcNatGateway has no Ready condition — always Active
	states[constants.FieldCommonState] = constants.StateCommonActive

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKubeOVNVpcNatGateway,
		States:       states,
	}, nil
}
