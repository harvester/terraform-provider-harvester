package importer

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceKubeOVNOvnFipStateGetter(obj *kubeovnv1.OvnFip) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:               obj.Name,
		constants.FieldCommonDescription:        GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:               GetTags(obj.Labels),
		constants.FieldCommonLabels:             GetLabels(obj.Labels),
		constants.FieldKubeOVNOvnFipOvnEip:      obj.Spec.OvnEip,
		constants.FieldKubeOVNOvnFipIPType:      obj.Spec.IPType,
		constants.FieldKubeOVNOvnFipIPName:      obj.Spec.IPName,
		constants.FieldKubeOVNOvnFipVpc:         obj.Spec.Vpc,
		constants.FieldKubeOVNOvnFipV4IP:        obj.Spec.V4Ip,
		constants.FieldKubeOVNOvnFipV6IP:        obj.Spec.V6Ip,
		constants.FieldKubeOVNOvnFipStatusReady: obj.Status.Ready,
		constants.FieldKubeOVNOvnFipStatusV4Eip: obj.Status.V4Eip,
		constants.FieldKubeOVNOvnFipStatusV6Eip: obj.Status.V6Eip,
		constants.FieldKubeOVNOvnFipStatusV4IP:  obj.Status.V4Ip,
		constants.FieldKubeOVNOvnFipStatusV6IP:  obj.Status.V6Ip,
		constants.FieldKubeOVNOvnFipStatusVpc:   obj.Status.Vpc,
	}

	if obj.Status.Ready {
		states[constants.FieldCommonState] = constants.StateCommonReady
	} else {
		states[constants.FieldCommonState] = constants.StateCommonActive
	}

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKubeOVNOvnFip,
		States:       states,
	}, nil
}
