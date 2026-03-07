package kubeovn_vip

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var _ util.Constructor = &Constructor{}

type Constructor struct {
	Vip *kubeovnv1.Vip
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.Vip.Labels).
		Labels(&c.Vip.Labels).
		Description(&c.Vip.Annotations).
		String(constants.FieldKubeOVNVipNamespace, &c.Vip.Spec.Namespace, false).
		String(constants.FieldKubeOVNVipSubnet, &c.Vip.Spec.Subnet, true).
		String(constants.FieldKubeOVNVipType, &c.Vip.Spec.Type, false).
		String(constants.FieldKubeOVNVipV4IP, &c.Vip.Spec.V4ip, false).
		String(constants.FieldKubeOVNVipV6IP, &c.Vip.Spec.V6ip, false).
		String(constants.FieldKubeOVNVipMacAddress, &c.Vip.Spec.MacAddress, false).
		String(constants.FieldKubeOVNVipParentV4IP, &c.Vip.Spec.ParentV4ip, false).
		String(constants.FieldKubeOVNVipParentV6IP, &c.Vip.Spec.ParentV6ip, false).
		String(constants.FieldKubeOVNVipParentMac, &c.Vip.Spec.ParentMac, false)

	customProcessors := []util.Processor{
		{
			Field: constants.FieldKubeOVNVipSelector,
			Parser: func(i interface{}) error {
				c.Vip.Spec.Selector = append(c.Vip.Spec.Selector, i.(string))
				return nil
			},
		},
		{
			Field: constants.FieldKubeOVNVipAttachSubnets,
			Parser: func(i interface{}) error {
				c.Vip.Spec.AttachSubnets = append(c.Vip.Spec.AttachSubnets, i.(string))
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
	return c.Vip, nil
}

func Creator(name string) util.Constructor {
	return &Constructor{
		Vip: &kubeovnv1.Vip{
			ObjectMeta: util.NewObjectMeta("", name),
		},
	}
}

func Updater(obj *kubeovnv1.Vip) util.Constructor {
	obj.Spec.Selector = nil
	obj.Spec.AttachSubnets = nil
	return &Constructor{Vip: obj}
}
