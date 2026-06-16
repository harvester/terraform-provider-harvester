package importer

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceKubeOVNSubnetStateGetter(obj *kubeovnv1.Subnet) (*StateGetter, error) {
	enableLb := true
	if obj.Spec.EnableLb != nil {
		enableLb = *obj.Spec.EnableLb
	}

	states := map[string]interface{}{
		constants.FieldCommonName:                  obj.Name,
		constants.FieldCommonDescription:           GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                  GetTags(obj.Labels),
		constants.FieldCommonLabels:                GetLabels(obj.Labels),
		constants.FieldKubeOVNSubnetVpc:            obj.Spec.Vpc,
		constants.FieldKubeOVNSubnetCIDRBlock:      obj.Spec.CIDRBlock,
		constants.FieldKubeOVNSubnetGateway:        obj.Spec.Gateway,
		constants.FieldKubeOVNSubnetExcludeIPs:     obj.Spec.ExcludeIps,
		constants.FieldKubeOVNSubnetProtocol:       obj.Spec.Protocol,
		constants.FieldKubeOVNSubnetVlan:           obj.Spec.Vlan,
		constants.FieldKubeOVNSubnetProvider:       obj.Spec.Provider,
		constants.FieldKubeOVNSubnetNamespaces:     obj.Spec.Namespaces,
		constants.FieldKubeOVNSubnetEnableDHCP:     obj.Spec.EnableDHCP,
		constants.FieldKubeOVNSubnetDHCPv4Options:  obj.Spec.DHCPv4Options,
		constants.FieldKubeOVNSubnetPrivate:        obj.Spec.Private,
		constants.FieldKubeOVNSubnetAllowSubnets:   obj.Spec.AllowSubnets,
		constants.FieldKubeOVNSubnetNatOutgoing:    obj.Spec.NatOutgoing,
		constants.FieldKubeOVNSubnetGatewayType:    obj.Spec.GatewayType,
		constants.FieldKubeOVNSubnetGatewayNode:    obj.Spec.GatewayNode,
		constants.FieldKubeOVNSubnetEnableLb:       enableLb,
		constants.FieldKubeOVNSubnetV4AvailableIPs: obj.Status.V4AvailableIPs,
		constants.FieldKubeOVNSubnetV4UsingIPs:     obj.Status.V4UsingIPs,
	}

	states[constants.FieldCommonState] = constants.StateCommonActive
	for _, condition := range obj.Status.Conditions {
		if kubeovnv1.ConditionType(condition.Type) == "Ready" && condition.Status == corev1.ConditionStatus("True") {
			states[constants.FieldCommonState] = constants.StateCommonReady
		}
	}

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKubeOVNSubnet,
		States:       states,
	}, nil
}
