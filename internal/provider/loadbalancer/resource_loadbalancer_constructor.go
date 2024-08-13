package loadbalancer

import (
	loadbalancerv1 "github.com/harvester/harvester-load-balancer/pkg/apis/loadbalancer.harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	LoadBalancer *loadbalancerv1.LoadBalancer
}

func (c *Constructor) Setup() util.Processors {
	return util.NewProcessors().
		Tags(&c.LoadBalancer.Labels).
		Description(&c.LoadBalancer.Annotations).
		String(constants.FieldLoadBalancerDescription, &c.LoadBalancer.Spec.Description, false)
}

func (c *Constructor) Validate() error {
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.LoadBalancer, nil
}

func newLoadBalancerConstructor(loadbalancer *loadbalancerv1.LoadBalancer) util.Constructor {
	return &Constructor{
		LoadBalancer: loadbalancer,
	}
}

func Creator(namespace, name string) util.Constructor {
	loadbalancer := &loadbalancerv1.LoadBalancer{
		ObjectMeta: util.NewObjectMeta(namespace, name),
	}
	return newLoadBalancerConstructor(loadbalancer)
}

func Updater(loadbalancer *loadbalancerv1.LoadBalancer) util.Constructor {
	return newLoadBalancerConstructor(loadbalancer)
}
