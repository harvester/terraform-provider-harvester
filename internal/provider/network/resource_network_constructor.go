package network

import (
	"errors"
	"fmt"

	nadv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/builder"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	Network *nadv1.NetworkAttachmentDefinition
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().Tags(&c.Network.Labels).Description(&c.Network.Annotations)
	customProcessors := []util.Processor{
		{
			Field: constants.FieldNetworkVlanID,
			Parser: func(i interface{}) error {
				var (
					networkType string
					vlanID      = i.(int)
				)
				if vlanID != 0 {
					networkType = builder.NetworkTypeVLAN
					c.Network.Spec.Config = fmt.Sprintf(builder.NetworkVLANConfigTemplate, c.Network.Name, vlanID)
				} else {
					networkType = builder.NetworkTypeCustom
				}
				c.Network.Labels = map[string]string{
					builder.LabelKeyNetworkType: networkType,
				}
				return nil
			},
			Required: true,
		},
		{
			Field: constants.FieldNetworkConfig,
			Parser: func(i interface{}) error {
				if c.Network.Labels[builder.LabelKeyNetworkType] == builder.NetworkTypeVLAN {
					return nil
				}
				config := i.(string)
				if config == "" {
					return errors.New("must specify config in custom network type")
				}
				c.Network.Spec.Config = config
				return nil
			},
		},
	}
	return append(processors, customProcessors...)
}

func (c *Constructor) Result() (interface{}, error) {
	return c.Network, nil
}

func newNetworkConstructor(network *nadv1.NetworkAttachmentDefinition) util.Constructor {
	return &Constructor{
		Network: network,
	}
}

func Creator(namespace, name string) util.Constructor {
	Network := &nadv1.NetworkAttachmentDefinition{
		ObjectMeta: util.NewObjectMeta(namespace, name),
	}
	return newNetworkConstructor(Network)
}

func Updater(network *nadv1.NetworkAttachmentDefinition) util.Constructor {
	return newNetworkConstructor(network)
}
