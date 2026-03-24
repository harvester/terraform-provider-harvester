package importer

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceKubeOVNOvnEipStateGetter(obj *kubeovnv1.OvnEip) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:                  obj.Name,
		constants.FieldCommonDescription:           GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                  GetTags(obj.Labels),
		constants.FieldCommonLabels:                GetLabels(obj.Labels),
		constants.FieldKubeOVNOvnEipExternalSubnet: obj.Spec.ExternalSubnet,
		constants.FieldKubeOVNOvnEipV4IP:           obj.Spec.V4Ip,
		constants.FieldKubeOVNOvnEipV6IP:           obj.Spec.V6Ip,
		constants.FieldKubeOVNOvnEipMacAddress:     obj.Spec.MacAddress,
		constants.FieldKubeOVNOvnEipType:           obj.Spec.Type,
		constants.FieldKubeOVNOvnEipStatusReady:    obj.Status.Ready,
		constants.FieldKubeOVNOvnEipStatusV4IP:     obj.Status.V4Ip,
		constants.FieldKubeOVNOvnEipStatusV6IP:     obj.Status.V6Ip,
		constants.FieldKubeOVNOvnEipStatusMac:      obj.Status.MacAddress,
		constants.FieldKubeOVNOvnEipStatusNat:      obj.Status.Nat,
		constants.FieldKubeOVNOvnEipStatusType:     obj.Status.Type,
	}

	if obj.Status.Ready {
		states[constants.FieldCommonState] = constants.StateCommonReady
	} else {
		states[constants.FieldCommonState] = constants.StateCommonActive
	}

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKubeOVNOvnEip,
		States:       states,
	}, nil
}
