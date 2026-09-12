package importer

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceKubeOVNOvnDnatRuleStateGetter(obj *kubeovnv1.OvnDnatRule) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:                 obj.Name,
		constants.FieldCommonDescription:          GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                 GetTags(obj.Labels),
		constants.FieldCommonLabels:               GetLabels(obj.Labels),
		constants.FieldKubeOVNOvnDnatOvnEip:       obj.Spec.OvnEip,
		constants.FieldKubeOVNOvnDnatIPType:       obj.Spec.IPType,
		constants.FieldKubeOVNOvnDnatIPName:       obj.Spec.IPName,
		constants.FieldKubeOVNOvnDnatInternalPort: obj.Spec.InternalPort,
		constants.FieldKubeOVNOvnDnatExternalPort: obj.Spec.ExternalPort,
		constants.FieldKubeOVNOvnDnatProtocol:     obj.Spec.Protocol,
		constants.FieldKubeOVNOvnDnatVpc:          obj.Spec.Vpc,
		constants.FieldKubeOVNOvnDnatV4IP:         obj.Spec.V4Ip,
		constants.FieldKubeOVNOvnDnatV6IP:         obj.Spec.V6Ip,
		constants.FieldKubeOVNOvnDnatStatusReady:  obj.Status.Ready,
		constants.FieldKubeOVNOvnDnatStatusV4Eip:  obj.Status.V4Eip,
		constants.FieldKubeOVNOvnDnatStatusV6Eip:  obj.Status.V6Eip,
		constants.FieldKubeOVNOvnDnatStatusVpc:    obj.Status.Vpc,
		constants.FieldKubeOVNOvnDnatStatusIPName: obj.Status.IPName,
	}

	if obj.Status.Ready {
		states[constants.FieldCommonState] = constants.StateCommonReady
	} else {
		states[constants.FieldCommonState] = constants.StateCommonActive
	}

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKubeOVNOvnDnatRule,
		States:       states,
	}, nil
}
