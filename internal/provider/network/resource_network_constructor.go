package network

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	networkutils "github.com/harvester/harvester-network-controller/pkg/utils"
	"github.com/harvester/harvester/pkg/builder"
	nadv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	Client  *client.Client
	Context context.Context

	Network           *nadv1.NetworkAttachmentDefinition
	Layer3NetworkConf *networkutils.Layer3NetworkConf
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().Tags(&c.Network.Labels).Description(&c.Network.Annotations)
	customProcessors := []util.Processor{
		{
			Field: constants.FieldNetworkClusterNetworkName,
			Parser: func(i interface{}) error {
				clusterNetworkName := i.(string)
				c.Network.Labels[networkutils.KeyClusterNetworkLabel] = clusterNetworkName
				return nil
			},
		},
		{
			Field: constants.FieldNetworkVlanID,
			Parser: func(i interface{}) error {
				var (
					networkType string
					vlanID      = i.(int)
				)
				if vlanID != 0 {
					networkType = builder.NetworkTypeVLAN
					clusterNetworkName := c.Network.Labels[networkutils.KeyClusterNetworkLabel]
					c.Network.Spec.Config = fmt.Sprintf(builder.NetworkVLANConfigTemplate, c.Network.Name, clusterNetworkName, vlanID)
				} else {
					networkType = builder.NetworkTypeCustom
				}
				c.Network.Labels[networkutils.KeyVlanLabel] = strconv.Itoa(vlanID)
				c.Network.Labels[builder.LabelKeyNetworkType] = networkType
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
			Required: true,
		},
		{
			Field: constants.FieldNetworkRouteDHCPServerIP,
			Parser: func(i interface{}) error {
				c.Layer3NetworkConf.ServerIPAddr = i.(string)
				return nil
			},
		},
		{
			Field: constants.FieldNetworkRouteCIDR,
			Parser: func(i interface{}) error {
				c.Layer3NetworkConf.CIDR = i.(string)
				return nil
			},
		},
		{
			Field: constants.FieldNetworkRouteGateWay,
			Parser: func(i interface{}) error {
				c.Layer3NetworkConf.Gateway = i.(string)
				return nil
			},
		},
		{
			Field: constants.FieldNetworkRouteMode,
			Parser: func(i interface{}) error {
				c.Layer3NetworkConf.Mode = networkutils.Mode(i.(string))
				layer3NetworkConf, err := c.Layer3NetworkConf.ToString()
				if err != nil {
					return err
				}
				switch c.Layer3NetworkConf.Mode {
				case networkutils.Manual:
					if c.Layer3NetworkConf.CIDR == "" {
						return errors.New("must specify route_cidr in manual route type")
					}
					if c.Layer3NetworkConf.Gateway == "" {
						return errors.New("must specify route_gateway in manual route type")
					}
				case networkutils.Auto:
					if c.Layer3NetworkConf.CIDR != "" {
						return errors.New("can not use route_mode auto when route_cidr has been specified")
					}
					if c.Layer3NetworkConf.Gateway != "" {
						return errors.New("can not use route_mode auto when route_gateway has been specified")
					}
				}
				if _, err = networkutils.NewLayer3NetworkConf(layer3NetworkConf); err != nil {
					return err
				}
				c.Network.Annotations[networkutils.KeyNetworkRoute] = layer3NetworkConf
				return nil
			},
			Required: true,
		},
	}
	return append(processors, customProcessors...)
}

func (c *Constructor) Validate() error {
	clusterNetworkName := c.Network.Labels[networkutils.KeyClusterNetworkLabel]
	if err := c.waitForClusterNetworkReady(clusterNetworkName, 1*time.Minute); err != nil {
		return fmt.Errorf("can not use the unready clusternetwork %s in networks, err: %v", clusterNetworkName, err)
	}
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.Network, nil
}

func newNetworkConstructor(c *client.Client, ctx context.Context, network *nadv1.NetworkAttachmentDefinition) util.Constructor {
	return &Constructor{
		Client:            c,
		Context:           ctx,
		Network:           network,
		Layer3NetworkConf: &networkutils.Layer3NetworkConf{},
	}
}

func Creator(c *client.Client, ctx context.Context, namespace, name string) util.Constructor {
	Network := &nadv1.NetworkAttachmentDefinition{
		ObjectMeta: util.NewObjectMeta(namespace, name),
	}
	return newNetworkConstructor(c, ctx, Network)
}

func Updater(c *client.Client, ctx context.Context, network *nadv1.NetworkAttachmentDefinition) util.Constructor {
	return newNetworkConstructor(c, ctx, network)
}
