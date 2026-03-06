package kubeovn_vlan

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var _ util.Constructor = &Constructor{}

type Constructor struct {
	Vlan *kubeovnv1.Vlan
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.Vlan.Labels).
		Labels(&c.Vlan.Labels).
		Description(&c.Vlan.Annotations).
		String(constants.FieldKubeOVNVlanProvider, &c.Vlan.Spec.Provider, false)

	customProcessors := []util.Processor{
		{
			Field: constants.FieldKubeOVNVlanID,
			Parser: func(i interface{}) error {
				c.Vlan.Spec.ID = i.(int)
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
	return c.Vlan, nil
}

func Creator(name string) util.Constructor {
	return &Constructor{
		Vlan: &kubeovnv1.Vlan{
			ObjectMeta: util.NewObjectMeta("", name),
		},
	}
}

func Updater(obj *kubeovnv1.Vlan) util.Constructor {
	return &Constructor{Vlan: obj}
}
