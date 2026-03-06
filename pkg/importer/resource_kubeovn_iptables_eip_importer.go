package importer

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceKubeOVNIptablesEIPStateGetter(obj *kubeovnv1.IptablesEIP) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:                       obj.Name,
		constants.FieldCommonDescription:                GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                       GetTags(obj.Labels),
		constants.FieldCommonLabels:                     GetLabels(obj.Labels),
		constants.FieldKubeOVNIptablesEIPV4IP:           obj.Spec.V4ip,
		constants.FieldKubeOVNIptablesEIPV6IP:           obj.Spec.V6ip,
		constants.FieldKubeOVNIptablesEIPMacAddress:     obj.Spec.MacAddress,
		constants.FieldKubeOVNIptablesEIPNatGwDp:        obj.Spec.NatGwDp,
		constants.FieldKubeOVNIptablesEIPQoSPolicy:      obj.Spec.QoSPolicy,
		constants.FieldKubeOVNIptablesEIPExternalSubnet: obj.Spec.ExternalSubnet,
		constants.FieldKubeOVNIptablesEIPReady:          obj.Status.Ready,
		constants.FieldKubeOVNIptablesEIPStatusIP:       obj.Status.IP,
		constants.FieldKubeOVNIptablesEIPStatusNat:      obj.Status.Nat,
		constants.FieldKubeOVNIptablesEIPStatusQoS:      obj.Status.QoSPolicy,
	}

	if obj.Status.Ready {
		states[constants.FieldCommonState] = constants.StateCommonReady
	} else {
		states[constants.FieldCommonState] = constants.StateCommonActive
	}

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKubeOVNIptablesEIP,
		States:       states,
	}, nil
}
