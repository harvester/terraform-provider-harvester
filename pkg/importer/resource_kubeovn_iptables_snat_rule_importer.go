package importer

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceKubeOVNIptablesSnatRuleStateGetter(obj *kubeovnv1.IptablesSnatRule) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:                      obj.Name,
		constants.FieldCommonDescription:               GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                      GetTags(obj.Labels),
		constants.FieldCommonLabels:                    GetLabels(obj.Labels),
		constants.FieldKubeOVNIptablesSnatEIP:          obj.Spec.EIP,
		constants.FieldKubeOVNIptablesSnatInternalCIDR: obj.Spec.InternalCIDR,
		constants.FieldKubeOVNIptablesSnatReady:        obj.Status.Ready,
		constants.FieldKubeOVNIptablesSnatStatusV4IP:   obj.Status.V4ip,
		constants.FieldKubeOVNIptablesSnatStatusV6IP:   obj.Status.V6ip,
		constants.FieldKubeOVNIptablesSnatStatusNat:    obj.Status.NatGwDp,
		constants.FieldKubeOVNIptablesSnatStatusCIDR:   obj.Status.InternalCIDR,
	}

	if obj.Status.Ready {
		states[constants.FieldCommonState] = constants.StateCommonReady
	} else {
		states[constants.FieldCommonState] = constants.StateCommonActive
	}

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKubeOVNIptablesSnatRule,
		States:       states,
	}, nil
}
