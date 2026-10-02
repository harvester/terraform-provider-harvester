package importer

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceKubeOVNIptablesFIPRuleStateGetter(obj *kubeovnv1.IptablesFIPRule) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:                   obj.Name,
		constants.FieldCommonDescription:            GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                   GetTags(obj.Labels),
		constants.FieldCommonLabels:                 GetLabels(obj.Labels),
		constants.FieldKubeOVNIptablesFIPEIP:        obj.Spec.EIP,
		constants.FieldKubeOVNIptablesFIPInternalIP: obj.Spec.InternalIP,
		constants.FieldKubeOVNIptablesFIPReady:      obj.Status.Ready,
		constants.FieldKubeOVNIptablesFIPStatusV4IP: obj.Status.V4ip,
		constants.FieldKubeOVNIptablesFIPStatusV6IP: obj.Status.V6ip,
		constants.FieldKubeOVNIptablesFIPStatusNat:  obj.Status.NatGwDp,
		constants.FieldKubeOVNIptablesFIPStatusIP:   obj.Status.InternalIP,
	}

	if obj.Status.Ready {
		states[constants.FieldCommonState] = constants.StateCommonReady
	} else {
		states[constants.FieldCommonState] = constants.StateCommonActive
	}

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKubeOVNIptablesFIPRule,
		States:       states,
	}, nil
}
