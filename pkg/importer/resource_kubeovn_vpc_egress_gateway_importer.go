package importer

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func flattenEgressPolicies(policies []kubeovnv1.VpcEgressGatewayPolicy) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(policies))
	for _, p := range policies {
		result = append(result, map[string]interface{}{
			constants.FieldKubeOVNVpcEgressGatewayPolicySNAT:     p.SNAT,
			constants.FieldKubeOVNVpcEgressGatewayPolicyIPBlocks: p.IPBlocks,
			constants.FieldKubeOVNVpcEgressGatewayPolicySubnets:  p.Subnets,
		})
	}
	return result
}

func flattenBFDConfig(bfd kubeovnv1.VpcEgressGatewayBFDConfig) []map[string]interface{} {
	if !bfd.Enabled {
		return nil
	}
	return []map[string]interface{}{
		{
			constants.FieldKubeOVNVpcEgressGatewayBFDEnabled:    bfd.Enabled,
			constants.FieldKubeOVNVpcEgressGatewayBFDMinRX:      int(bfd.MinRX),
			constants.FieldKubeOVNVpcEgressGatewayBFDMinTX:      int(bfd.MinTX),
			constants.FieldKubeOVNVpcEgressGatewayBFDMultiplier: int(bfd.Multiplier),
		},
	}
}

func ResourceKubeOVNVpcEgressGatewayStateGetter(obj *kubeovnv1.VpcEgressGateway) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonNamespace:                          obj.Namespace,
		constants.FieldCommonName:                               obj.Name,
		constants.FieldCommonDescription:                        GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                               GetTags(obj.Labels),
		constants.FieldCommonLabels:                             GetLabels(obj.Labels),
		constants.FieldKubeOVNVpcEgressGatewayVpc:               obj.Spec.VPC,
		constants.FieldKubeOVNVpcEgressGatewayReplicas:          int(obj.Spec.Replicas),
		constants.FieldKubeOVNVpcEgressGatewayPrefix:            obj.Spec.Prefix,
		constants.FieldKubeOVNVpcEgressGatewayImage:             obj.Spec.Image,
		constants.FieldKubeOVNVpcEgressGatewayInternalSubnet:    obj.Spec.InternalSubnet,
		constants.FieldKubeOVNVpcEgressGatewayExternalSubnet:    obj.Spec.ExternalSubnet,
		constants.FieldKubeOVNVpcEgressGatewayInternalIPs:       obj.Spec.InternalIPs,
		constants.FieldKubeOVNVpcEgressGatewayExternalIPs:       obj.Spec.ExternalIPs,
		constants.FieldKubeOVNVpcEgressGatewayTrafficPolicy:     obj.Spec.TrafficPolicy,
		constants.FieldKubeOVNVpcEgressGatewayBFD:               flattenBFDConfig(obj.Spec.BFD),
		constants.FieldKubeOVNVpcEgressGatewayPolicies:          flattenEgressPolicies(obj.Spec.Policies),
		constants.FieldKubeOVNVpcEgressGatewayStatusReady:       obj.Status.Ready,
		constants.FieldKubeOVNVpcEgressGatewayStatusPhase:       string(obj.Status.Phase),
		constants.FieldKubeOVNVpcEgressGatewayStatusInternalIPs: obj.Status.InternalIPs,
		constants.FieldKubeOVNVpcEgressGatewayStatusExternalIPs: obj.Status.ExternalIPs,
	}

	if obj.Status.Ready {
		states[constants.FieldCommonState] = constants.StateCommonReady
	} else {
		states[constants.FieldCommonState] = constants.StateCommonActive
	}

	return &StateGetter{
		ID:           helper.BuildID(obj.Namespace, obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKubeOVNVpcEgressGateway,
		States:       states,
	}, nil
}
