package kubeovn_vpc_nat_gateway

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var _ util.Constructor = &Constructor{}

type Constructor struct {
	Gateway *kubeovnv1.VpcNatGateway
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.Gateway.Labels).
		Labels(&c.Gateway.Labels).
		Description(&c.Gateway.Annotations).
		String(constants.FieldKubeOVNVpcNatGwVpc, &c.Gateway.Spec.Vpc, true).
		String(constants.FieldKubeOVNVpcNatGwSubnet, &c.Gateway.Spec.Subnet, true).
		String(constants.FieldKubeOVNVpcNatGwLanIP, &c.Gateway.Spec.LanIP, false).
		String(constants.FieldKubeOVNVpcNatGwQoSPolicy, &c.Gateway.Spec.QoSPolicy, false)

	customProcessors := []util.Processor{
		{
			Field: constants.FieldKubeOVNVpcNatGwExternalSubnets,
			Parser: func(i interface{}) error {
				c.Gateway.Spec.ExternalSubnets = append(c.Gateway.Spec.ExternalSubnets, i.(string))
				return nil
			},
		},
		{
			Field: constants.FieldKubeOVNVpcNatGwSelector,
			Parser: func(i interface{}) error {
				c.Gateway.Spec.Selector = append(c.Gateway.Spec.Selector, i.(string))
				return nil
			},
		},
	}
	return append(processors, customProcessors...)
}

func (c *Constructor) Validate() error {
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.Gateway, nil
}

func Creator(name string) util.Constructor {
	return &Constructor{
		Gateway: &kubeovnv1.VpcNatGateway{
			ObjectMeta: util.NewObjectMeta("", name),
		},
	}
}

func Updater(obj *kubeovnv1.VpcNatGateway) util.Constructor {
	obj.Spec.ExternalSubnets = nil
	obj.Spec.Selector = nil
	return &Constructor{Gateway: obj}
}
