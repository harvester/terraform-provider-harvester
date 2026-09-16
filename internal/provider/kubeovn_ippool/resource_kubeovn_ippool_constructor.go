package kubeovn_ippool

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var _ util.Constructor = &Constructor{}

type Constructor struct {
	IPPool *kubeovnv1.IPPool
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.IPPool.Labels).
		Labels(&c.IPPool.Labels).
		Description(&c.IPPool.Annotations).
		String(constants.FieldKubeOVNIPPoolSubnet, &c.IPPool.Spec.Subnet, true)

	customProcessors := []util.Processor{
		{
			Field: constants.FieldKubeOVNIPPoolIPs,
			Parser: func(i interface{}) error {
				c.IPPool.Spec.IPs = append(c.IPPool.Spec.IPs, i.(string))
				return nil
			},
		},
		{
			Field: constants.FieldKubeOVNIPPoolNamespaces,
			Parser: func(i interface{}) error {
				c.IPPool.Spec.Namespaces = append(c.IPPool.Spec.Namespaces, i.(string))
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
	return c.IPPool, nil
}

func Creator(name string) util.Constructor {
	return &Constructor{
		IPPool: &kubeovnv1.IPPool{
			ObjectMeta: util.NewObjectMeta("", name),
		},
	}
}

func Updater(obj *kubeovnv1.IPPool) util.Constructor {
	obj.Spec.IPs = nil
	obj.Spec.Namespaces = nil
	return &Constructor{IPPool: obj}
}
