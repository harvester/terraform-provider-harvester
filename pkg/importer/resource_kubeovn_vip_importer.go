package importer

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceKubeOVNVipStateGetter(obj *kubeovnv1.Vip) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:              obj.Name,
		constants.FieldCommonDescription:       GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:              GetTags(obj.Labels),
		constants.FieldCommonLabels:            GetLabels(obj.Labels),
		constants.FieldKubeOVNVipNamespace:     obj.Spec.Namespace,
		constants.FieldKubeOVNVipSubnet:        obj.Spec.Subnet,
		constants.FieldKubeOVNVipType:          obj.Spec.Type,
		constants.FieldKubeOVNVipV4IP:          obj.Spec.V4ip,
		constants.FieldKubeOVNVipV6IP:          obj.Spec.V6ip,
		constants.FieldKubeOVNVipMacAddress:    obj.Spec.MacAddress,
		constants.FieldKubeOVNVipSelector:      obj.Spec.Selector,
		constants.FieldKubeOVNVipAttachSubnets: obj.Spec.AttachSubnets,
		constants.FieldKubeOVNVipStatusV4IP:    obj.Status.V4ip,
		constants.FieldKubeOVNVipStatusV6IP:    obj.Status.V6ip,
		constants.FieldKubeOVNVipStatusMac:     obj.Status.Mac,
		constants.FieldKubeOVNVipStatusType:    obj.Status.Type,
	}

	if obj.Status.V4ip != "" || obj.Status.V6ip != "" {
		states[constants.FieldCommonState] = constants.StateCommonReady
	} else {
		states[constants.FieldCommonState] = constants.StateCommonActive
	}

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKubeOVNVip,
		States:       states,
	}, nil
}
