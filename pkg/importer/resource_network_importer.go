package importer

import (
	"encoding/json"

	networkapi "github.com/harvester/harvester-network-controller/pkg/apis/network.harvesterhci.io"
	networkutils "github.com/harvester/harvester-network-controller/pkg/utils"
	"github.com/harvester/harvester/pkg/builder"
	nadv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceNetworkStateGetter(obj *nadv1.NetworkAttachmentDefinition) (*StateGetter, error) {
	var (
		vlanID            int
		networkType       string
		networkConf       string
		layer3NetworkConf = &networkutils.Layer3NetworkConf{}
		err               error
	)
	if obj.Labels != nil {
		networkType = obj.Labels[builder.LabelKeyNetworkType]
	}
	if networkType == builder.NetworkTypeVLAN {
		netconf := &networkutils.NetConf{}
		if err = json.Unmarshal([]byte(obj.Spec.Config), netconf); err != nil {
			return nil, err
		}
		vlanID = netconf.Vlan
	}
	if obj.Annotations != nil {
		networkConf = obj.Annotations[networkapi.GroupName+"/route"]
	}
	if networkConf != "" {
		layer3NetworkConf, err = networkutils.NewLayer3NetworkConf(networkConf)
		if err != nil {
			return nil, err
		}
	}

	states := map[string]interface{}{
		constants.FieldCommonNamespace:           obj.Namespace,
		constants.FieldCommonName:                obj.Name,
		constants.FieldCommonDescription:         GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                GetTags(obj.Labels),
		constants.FieldCommonLabels:              GetLabels(obj.Labels),
		constants.FieldNetworkVlanID:             vlanID,
		constants.FieldNetworkConfig:             obj.Spec.Config,
		constants.FieldNetworkRouteMode:          layer3NetworkConf.Mode,
		constants.FieldNetworkRouteDHCPServerIP:  layer3NetworkConf.ServerIPAddr,
		constants.FieldNetworkRouteCIDR:          layer3NetworkConf.CIDR,
		constants.FieldNetworkRouteGateWay:       layer3NetworkConf.Gateway,
		constants.FieldNetworkRouteConnectivity:  layer3NetworkConf.Connectivity,
		constants.FieldNetworkClusterNetworkName: obj.Labels[networkutils.KeyClusterNetworkLabel],
	}
	if layer3NetworkConf.Mode == networkutils.Manual {
		states[constants.FieldNetworkRouteCIDR] = layer3NetworkConf.CIDR
		states[constants.FieldNetworkRouteGateWay] = layer3NetworkConf.Gateway
	}
	return &StateGetter{
		ID:           helper.BuildID(obj.Namespace, obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeNetwork,
		States:       states,
	}, nil
}
