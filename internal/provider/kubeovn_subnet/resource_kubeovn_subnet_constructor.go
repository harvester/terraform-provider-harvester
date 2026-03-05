package kubeovn_subnet

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var _ util.Constructor = &Constructor{}

type Constructor struct {
	Subnet *kubeovnv1.Subnet
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.Subnet.Labels).
		Labels(&c.Subnet.Labels).
		Description(&c.Subnet.Annotations).
		String(constants.FieldKubeOVNSubnetVpc, &c.Subnet.Spec.Vpc, false).
		String(constants.FieldKubeOVNSubnetCIDRBlock, &c.Subnet.Spec.CIDRBlock, true).
		String(constants.FieldKubeOVNSubnetGateway, &c.Subnet.Spec.Gateway, true).
		String(constants.FieldKubeOVNSubnetProtocol, &c.Subnet.Spec.Protocol, false).
		String(constants.FieldKubeOVNSubnetVlan, &c.Subnet.Spec.Vlan, false).
		String(constants.FieldKubeOVNSubnetProvider, &c.Subnet.Spec.Provider, false).
		Bool(constants.FieldKubeOVNSubnetEnableDHCP, &c.Subnet.Spec.EnableDHCP, false).
		String(constants.FieldKubeOVNSubnetDHCPv4Options, &c.Subnet.Spec.DHCPv4Options, false).
		Bool(constants.FieldKubeOVNSubnetPrivate, &c.Subnet.Spec.Private, false).
		Bool(constants.FieldKubeOVNSubnetNatOutgoing, &c.Subnet.Spec.NatOutgoing, false).
		String(constants.FieldKubeOVNSubnetGatewayType, &c.Subnet.Spec.GatewayType, false).
		String(constants.FieldKubeOVNSubnetGatewayNode, &c.Subnet.Spec.GatewayNode, false)

	customProcessors := []util.Processor{
		{
			Field: constants.FieldKubeOVNSubnetExcludeIPs,
			Parser: func(i interface{}) error {
				c.Subnet.Spec.ExcludeIps = append(c.Subnet.Spec.ExcludeIps, i.(string))
				return nil
			},
		},
		{
			Field: constants.FieldKubeOVNSubnetNamespaces,
			Parser: func(i interface{}) error {
				c.Subnet.Spec.Namespaces = append(c.Subnet.Spec.Namespaces, i.(string))
				return nil
			},
		},
		{
			Field: constants.FieldKubeOVNSubnetAllowSubnets,
			Parser: func(i interface{}) error {
				c.Subnet.Spec.AllowSubnets = append(c.Subnet.Spec.AllowSubnets, i.(string))
				return nil
			},
		},
		{
			Field: constants.FieldKubeOVNSubnetEnableLb,
			Parser: func(i interface{}) error {
				val := i.(bool)
				c.Subnet.Spec.EnableLb = &val
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
	return c.Subnet, nil
}

func Creator(name string) util.Constructor {
	subnet := &kubeovnv1.Subnet{
		ObjectMeta: util.NewObjectMeta("", name),
	}
	return &Constructor{Subnet: subnet}
}

func Updater(subnet *kubeovnv1.Subnet) util.Constructor {
	subnet.Spec.ExcludeIps = nil
	subnet.Spec.Namespaces = nil
	subnet.Spec.AllowSubnets = nil
	return &Constructor{Subnet: subnet}
}
