package kubeovn_vpc_dns

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var _ util.Constructor = &Constructor{}

type Constructor struct {
	VpcDns *kubeovnv1.VpcDns
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.VpcDns.Labels).
		Labels(&c.VpcDns.Labels).
		Description(&c.VpcDns.Annotations).
		String(constants.FieldKubeOVNVpcDnsVpc, &c.VpcDns.Spec.Vpc, true).
		String(constants.FieldKubeOVNVpcDnsSubnet, &c.VpcDns.Spec.Subnet, true)

	customProcessors := []util.Processor{
		{
			Field:    constants.FieldKubeOVNVpcDnsReplicas,
			Required: true,
			Parser: func(i interface{}) error {
				c.VpcDns.Spec.Replicas = int32(i.(int))
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
	return c.VpcDns, nil
}

func Creator(name string) util.Constructor {
	return &Constructor{
		VpcDns: &kubeovnv1.VpcDns{
			ObjectMeta: util.NewObjectMeta("", name),
		},
	}
}

func Updater(obj *kubeovnv1.VpcDns) util.Constructor {
	return &Constructor{VpcDns: obj}
}
