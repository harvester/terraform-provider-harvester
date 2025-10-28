package importer

import (
	"encoding/json"

	harvsternetworkv1 "github.com/harvester/harvester-network-controller/pkg/apis/network.harvesterhci.io/v1beta1"
	harvsternetworkutils "github.com/harvester/harvester-network-controller/pkg/utils"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceVLANConfigStateGetter(obj *harvsternetworkv1.VlanConfig) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:                   obj.Name,
		constants.FieldCommonDescription:            GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                   GetTags(obj.Labels),
		constants.FieldCommonLabels:                 GetLabels(obj.Labels),
		constants.FieldVLANConfigClusterNetworkName: obj.Spec.ClusterNetwork,
		constants.FieldVLANConfigNodeSelector:       obj.Spec.NodeSelector,
	}

	// matchedNodes
	matchedNodes := []string{}
	if matchedNodesAnnotation := obj.Annotations[harvsternetworkutils.KeyMatchedNodes]; matchedNodesAnnotation != "" {
		if err := json.Unmarshal([]byte(matchedNodesAnnotation), &matchedNodes); err != nil {
			return nil, err
		}
	}
	states[constants.FieldVLANConfigMatchedNodes] = matchedNodes

	// uplink
	uplink := map[string]interface{}{}
	uplink[constants.FieldUplinkNICs] = obj.Spec.Uplink.NICs
	if bondOptions := obj.Spec.Uplink.BondOptions; bondOptions != nil {
		uplink[constants.FieldUplinkBondMode] = bondOptions.Mode
		uplink[constants.FieldUplinkBondMiimon] = bondOptions.Miimon
	}
	if linkAttrs := obj.Spec.Uplink.LinkAttrs; linkAttrs != nil {
		uplink[constants.FieldUplinkMTU] = linkAttrs.MTU
	}
	states[constants.FieldVLANConfigUplink] = []map[string]interface{}{uplink}

	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeVLANConfig,
		States:       states,
	}, nil
}
