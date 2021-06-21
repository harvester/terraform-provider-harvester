package importer

import (
	"encoding/json"

	"github.com/harvester/harvester/pkg/builder"
	"github.com/harvester/harvester/pkg/webhook/resources/network"
	nadv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceNetworkStateGetter(obj *nadv1.NetworkAttachmentDefinition) (*StateGetter, error) {
	var (
		vlanID      interface{}
		networkType = obj.Labels[builder.LabelKeyNetworkType]
	)
	if networkType == builder.NetworkTypeVLAN {
		netconf := &network.NetConf{}
		if err := json.Unmarshal([]byte(obj.Spec.Config), netconf); err != nil {
			return nil, err
		}
		vlanID = netconf.Vlan
	}
	states := map[string]interface{}{
		constants.FieldCommonNamespace:   obj.Namespace,
		constants.FieldCommonName:        obj.Name,
		constants.FieldCommonDescription: GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:        GetTags(obj.Labels),
		constants.FieldNetworkVlanID:     vlanID,
		constants.FieldNetworkConfig:     obj.Spec.Config,
	}
	return &StateGetter{
		ID:           helper.BuildID(obj.Namespace, obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeNetwork,
		States:       states,
	}, nil
}
