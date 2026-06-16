package importer

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceKubeOVNIptablesDnatRuleStateGetter(obj *kubeovnv1.IptablesDnatRule) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:                      obj.Name,
		constants.FieldCommonDescription:               GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                      GetTags(obj.Labels),
		constants.FieldCommonLabels:                    GetLabels(obj.Labels),
		constants.FieldKubeOVNIptablesDnatEIP:          obj.Spec.EIP,
		constants.FieldKubeOVNIptablesDnatExternalPort: obj.Spec.ExternalPort,
		constants.FieldKubeOVNIptablesDnatProtocol:     obj.Spec.Protocol,
		constants.FieldKubeOVNIptablesDnatInternalIP:   obj.Spec.InternalIP,
		constants.FieldKubeOVNIptablesDnatInternalPort: obj.Spec.InternalPort,
		constants.FieldKubeOVNIptablesDnatReady:        obj.Status.Ready,
		constants.FieldKubeOVNIptablesDnatStatusV4IP:   obj.Status.V4ip,
		constants.FieldKubeOVNIptablesDnatStatusV6IP:   obj.Status.V6ip,
		constants.FieldKubeOVNIptablesDnatStatusNat:    obj.Status.NatGwDp,
		constants.FieldKubeOVNIptablesDnatStatusProto:  obj.Status.Protocol,
		constants.FieldKubeOVNIptablesDnatStatusIntIP:  obj.Status.InternalIP,
		constants.FieldKubeOVNIptablesDnatStatusIntP:   obj.Status.InternalPort,
		constants.FieldKubeOVNIptablesDnatStatusExtP:   obj.Status.ExternalPort,
	}

	if obj.Status.Ready {
		states[constants.FieldCommonState] = constants.StateCommonReady
	} else {
		states[constants.FieldCommonState] = constants.StateCommonActive
	}

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKubeOVNIptablesDnatRule,
		States:       states,
	}, nil
}
