package importer

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceKubeOVNOvnSnatRuleStateGetter(obj *kubeovnv1.OvnSnatRule) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:                obj.Name,
		constants.FieldCommonDescription:         GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                GetTags(obj.Labels),
		constants.FieldCommonLabels:              GetLabels(obj.Labels),
		constants.FieldKubeOVNOvnSnatOvnEip:      obj.Spec.OvnEip,
		constants.FieldKubeOVNOvnSnatVpcSubnet:   obj.Spec.VpcSubnet,
		constants.FieldKubeOVNOvnSnatIPName:      obj.Spec.IPName,
		constants.FieldKubeOVNOvnSnatVpc:         obj.Spec.Vpc,
		constants.FieldKubeOVNOvnSnatV4IpCidr:    obj.Spec.V4IpCidr,
		constants.FieldKubeOVNOvnSnatV6IpCidr:    obj.Spec.V6IpCidr,
		constants.FieldKubeOVNOvnSnatStatusReady: obj.Status.Ready,
		constants.FieldKubeOVNOvnSnatStatusV4Eip: obj.Status.V4Eip,
		constants.FieldKubeOVNOvnSnatStatusV6Eip: obj.Status.V6Eip,
		constants.FieldKubeOVNOvnSnatStatusVpc:   obj.Status.Vpc,
	}

	if obj.Status.Ready {
		states[constants.FieldCommonState] = constants.StateCommonReady
	} else {
		states[constants.FieldCommonState] = constants.StateCommonActive
	}

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKubeOVNOvnSnatRule,
		States:       states,
	}, nil
}
