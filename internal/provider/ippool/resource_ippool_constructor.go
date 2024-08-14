package ippool

import (
	loadbalancerv1 "github.com/harvester/harvester-load-balancer/pkg/apis/loadbalancer.harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	IPPool *loadbalancerv1.IPPool
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.IPPool.Labels).
		String(constants.FieldIPPoolDescription, &c.IPPool.Spec.Description, false)

	subresourceProcessors := []util.Processor{
		{
			Field:    constants.SubresourceTypeIPPoolRange,
			Parser:   c.subresourceIPPoolRangeParser,
			Required: true,
		},
		{
			Field:    constants.SubresourceTypeIPPoolSelector,
			Parser:   c.subresourceIPPoolSelectorParser,
			Required: false,
		},
	}

	return append(processors, subresourceProcessors...)
}

func (c *Constructor) Validate() error {
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.IPPool, nil
}

func newIPPoolConstructor(ippool *loadbalancerv1.IPPool) util.Constructor {
	return &Constructor{
		IPPool: ippool,
	}
}

func Creator(name string) util.Constructor {
	ippool := &loadbalancerv1.IPPool{
		ObjectMeta: util.NewObjectMeta("", name),
	}
	return newIPPoolConstructor(ippool)
}

func Updater(ippool *loadbalancerv1.IPPool) util.Constructor {
	ippool.Spec.Ranges = []loadbalancerv1.Range{}
	ippool.Spec.Selector = loadbalancerv1.Selector{}
	return newIPPoolConstructor(ippool)
}

func (c *Constructor) subresourceIPPoolRangeParser(data interface{}) error {
	ippoolRange := data.(map[string]interface{})
	start := ippoolRange[constants.FieldRangeStart].(string)
	end := ippoolRange[constants.FieldRangeEnd].(string)
	subnet := ippoolRange[constants.FieldRangeSubnet].(string)
	gateway := ippoolRange[constants.FieldRangeGateway].(string)

	c.IPPool.Spec.Ranges = append(c.IPPool.Spec.Ranges, loadbalancerv1.Range{
		RangeStart: start,
		RangeEnd:   end,
		Subnet:     subnet,
		Gateway:    gateway,
	})
	return nil
}

func (c *Constructor) subresourceIPPoolSelectorParser(data interface{}) error {
	ippoolSelector := data.(map[string]interface{})

	priority := uint32(ippoolSelector[constants.FieldSelectorPriority].(int))
	network := ippoolSelector[constants.FieldSelectorNetwork].(string)

	scopesData := ippoolSelector[constants.SubresourceTypeIPPoolSelectorScope].([]interface{})

	scopes := []loadbalancerv1.Tuple{}
	for _, scopeData := range scopesData {
		scope := scopeData.(map[string]interface{})
		scopeProject := scope[constants.FieldScopeProject].(string)
		scopeNamespace := scope[constants.FieldScopeNamespace].(string)
		scopeGuestCluster := scope[constants.FieldScopeGuestCluster].(string)

		scopes = append(scopes, loadbalancerv1.Tuple{
			Project:      scopeProject,
			Namespace:    scopeNamespace,
			GuestCluster: scopeGuestCluster,
		})
	}

	c.IPPool.Spec.Selector = loadbalancerv1.Selector{
		Priority: priority,
		Network:  network,
		Scope:    scopes,
	}
	return nil
}
