package importer

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceKubeOVNIPPoolStateGetter(obj *kubeovnv1.IPPool) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:                    obj.Name,
		constants.FieldCommonDescription:             GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                    GetTags(obj.Labels),
		constants.FieldCommonLabels:                  GetLabels(obj.Labels),
		constants.FieldKubeOVNIPPoolSubnet:           obj.Spec.Subnet,
		constants.FieldKubeOVNIPPoolIPs:              obj.Spec.IPs,
		constants.FieldKubeOVNIPPoolNamespaces:       obj.Spec.Namespaces,
		constants.FieldKubeOVNIPPoolV4AvailableIPs:   obj.Status.V4AvailableIPs.String(),
		constants.FieldKubeOVNIPPoolV4AvailableRange: obj.Status.V4AvailableIPRange,
		constants.FieldKubeOVNIPPoolV4UsingIPs:       obj.Status.V4UsingIPs.String(),
		constants.FieldKubeOVNIPPoolV4UsingRange:     obj.Status.V4UsingIPRange,
		constants.FieldKubeOVNIPPoolV6AvailableIPs:   obj.Status.V6AvailableIPs.String(),
		constants.FieldKubeOVNIPPoolV6AvailableRange: obj.Status.V6AvailableIPRange,
		constants.FieldKubeOVNIPPoolV6UsingIPs:       obj.Status.V6UsingIPs.String(),
		constants.FieldKubeOVNIPPoolV6UsingRange:     obj.Status.V6UsingIPRange,
	}

	states[constants.FieldCommonState] = constants.StateCommonActive

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKubeOVNIPPool,
		States:       states,
	}, nil
}
