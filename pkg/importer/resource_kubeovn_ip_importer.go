package importer

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceKubeOVNIPStateGetter(obj *kubeovnv1.IP) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:          obj.Name,
		constants.FieldCommonDescription:   GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:          GetTags(obj.Labels),
		constants.FieldCommonLabels:        GetLabels(obj.Labels),
		constants.FieldKubeOVNIPPodName:    obj.Spec.PodName,
		constants.FieldKubeOVNIPNamespace:  obj.Spec.Namespace,
		constants.FieldKubeOVNIPSubnet:     obj.Spec.Subnet,
		constants.FieldKubeOVNIPIPAddress:  obj.Spec.IPAddress,
		constants.FieldKubeOVNIPMacAddress: obj.Spec.MacAddress,
		constants.FieldKubeOVNIPNodeName:   obj.Spec.NodeName,
		constants.FieldKubeOVNIPV4IP:       obj.Spec.V4IPAddress,
		constants.FieldKubeOVNIPV6IP:       obj.Spec.V6IPAddress,
	}
	states[constants.FieldCommonState] = constants.StateCommonActive

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKubeOVNIP,
		States:       states,
	}, nil
}
