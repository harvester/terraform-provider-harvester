package clusternetwork

import (
	"fmt"

	harvsternetworkv1 "github.com/harvester/harvester-network-controller/pkg/apis/network.harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	ClusterNetwork *harvsternetworkv1.ClusterNetwork
}

func (c *Constructor) Setup() util.Processors {
	if c.ClusterNetwork.Config == nil {
		c.ClusterNetwork.Config = map[string]string{}
	}
	processors := util.NewProcessors().Tags(&c.ClusterNetwork.Labels).Description(&c.ClusterNetwork.Annotations).
		Bool(constants.FieldClusterNetworkEnable, &c.ClusterNetwork.Enable, true)
	customProcessors := []util.Processor{
		{
			Field: constants.FieldClusterNetworkDefaultPhysicalNIC,
			Parser: func(i interface{}) error {
				defaultPhysicalNIC := i.(string)
				if c.ClusterNetwork.Enable && defaultPhysicalNIC == "" {
					return fmt.Errorf("%s is true, please specify %s", constants.FieldClusterNetworkEnable, constants.FieldClusterNetworkDefaultPhysicalNIC)
				}
				c.ClusterNetwork.Config[constants.ClusterNetworkConfigKeyDefaultPhysicalNIC] = defaultPhysicalNIC
				return nil
			},
			Required: true,
		},
	}
	return append(processors, customProcessors...)
}

func (c *Constructor) Validate() error {
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.ClusterNetwork, nil
}

func newClusterNetworkConstructor(clusterNetwork *harvsternetworkv1.ClusterNetwork) util.Constructor {
	return &Constructor{
		ClusterNetwork: clusterNetwork,
	}
}

func Creator(namespace, name string) util.Constructor {
	clusterNetwork := &harvsternetworkv1.ClusterNetwork{
		ObjectMeta: util.NewObjectMeta(namespace, name),
	}
	return newClusterNetworkConstructor(clusterNetwork)
}

func Updater(clusterNetwork *harvsternetworkv1.ClusterNetwork) util.Constructor {
	return newClusterNetworkConstructor(clusterNetwork)
}
