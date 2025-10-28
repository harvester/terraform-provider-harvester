package clusternetwork

import (
	harvsternetworkv1 "github.com/harvester/harvester-network-controller/pkg/apis/network.harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	ClusterNetwork *harvsternetworkv1.ClusterNetwork
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.ClusterNetwork.Labels).
		Labels(&c.ClusterNetwork.Labels).
		Description(&c.ClusterNetwork.Annotations)
	customProcessors := []util.Processor{}
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

func Creator(name string) util.Constructor {
	clusterNetwork := &harvsternetworkv1.ClusterNetwork{
		ObjectMeta: util.NewObjectMeta("", name),
	}
	return newClusterNetworkConstructor(clusterNetwork)
}

func Updater(clusterNetwork *harvsternetworkv1.ClusterNetwork) util.Constructor {
	return newClusterNetworkConstructor(clusterNetwork)
}
