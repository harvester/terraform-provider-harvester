package importer

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceKubeOVNVpcDnsStateGetter(obj *kubeovnv1.VpcDns) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:                obj.Name,
		constants.FieldCommonDescription:         GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                GetTags(obj.Labels),
		constants.FieldCommonLabels:              GetLabels(obj.Labels),
		constants.FieldKubeOVNVpcDnsReplicas:     int(obj.Spec.Replicas),
		constants.FieldKubeOVNVpcDnsVpc:          obj.Spec.Vpc,
		constants.FieldKubeOVNVpcDnsSubnet:       obj.Spec.Subnet,
		constants.FieldKubeOVNVpcDnsStatusActive: obj.Status.Active,
	}

	if obj.Status.Active {
		states[constants.FieldCommonState] = constants.StateCommonReady
	} else {
		states[constants.FieldCommonState] = constants.StateCommonActive
	}

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKubeOVNVpcDns,
		States:       states,
	}, nil
}
