package network

import (
	"context"
	"errors"
	"fmt"
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

	ClusterNetworkName string
	Network            *nadv1.NetworkAttachmentDefinition
	Layer3NetworkConf  *networkutils.Layer3NetworkConf
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().Tags(&c.Network.Labels).Description(&c.Network.Annotations)
	customProcessors := []util.Processor{
		{
			Field: constants.FieldNetworkClusterNetworkName,
			Parser: func(i interface{}) error {
				c.ClusterNetworkName = i.(string)
				c.Network.Labels[networkutils.KeyClusterNetworkLabel] = c.ClusterNetworkName
				return nil
			},
			Required: true,
		},
		{
			Field: constants.FieldNetworkVlanID,
			Parser: func(i interface{}) error {
				vlanID := i.(int)
				c.Network.Spec.Config = fmt.Sprintf(builder.NetworkVLANConfigTemplate, c.Network.Name, c.ClusterNetworkName, vlanID)

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
	if err := c.waitForClusterNetworkReady(c.ClusterNetworkName, 1*time.Minute); err != nil {
		return fmt.Errorf("can not use the unready clusternetwork %s in networks, err: %v", c.ClusterNetworkName, err)
	}
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.Network, nil
}

func newNetworkConstructor(c *client.Client, ctx context.Context, clusterNetworkName string, network *nadv1.NetworkAttachmentDefinition) util.Constructor {
	return &Constructor{
		Client:             c,
		Context:            ctx,
		ClusterNetworkName: clusterNetworkName,
		Network:            network,
		Layer3NetworkConf:  &networkutils.Layer3NetworkConf{},
	}
}

func Creator(c *client.Client, ctx context.Context, clusterNetworkName, namespace, name string) util.Constructor {
	network := &nadv1.NetworkAttachmentDefinition{
		ObjectMeta: util.NewObjectMeta(namespace, name),
	}
	network.Labels[networkutils.KeyClusterNetworkLabel] = clusterNetworkName
	return newNetworkConstructor(c, ctx, clusterNetworkName, network)
}

func Updater(c *client.Client, ctx context.Context, network *nadv1.NetworkAttachmentDefinition) util.Constructor {
	clusterNetworkName := network.Labels[networkutils.KeyClusterNetworkLabel]
	return newNetworkConstructor(c, ctx, clusterNetworkName, network)
}
