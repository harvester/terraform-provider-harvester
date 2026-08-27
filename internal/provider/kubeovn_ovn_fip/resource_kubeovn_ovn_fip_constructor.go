package kubeovn_ovn_fip

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var _ util.Constructor = &Constructor{}

type Constructor struct {
	OvnFip *kubeovnv1.OvnFip
}

func (c *Constructor) Setup() util.Processors {
	return util.NewProcessors().
		Tags(&c.OvnFip.Labels).
		Labels(&c.OvnFip.Labels).
		Description(&c.OvnFip.Annotations).
		String(constants.FieldKubeOVNOvnFipOvnEip, &c.OvnFip.Spec.OvnEip, true).
		String(constants.FieldKubeOVNOvnFipIPType, &c.OvnFip.Spec.IPType, false).
		String(constants.FieldKubeOVNOvnFipIPName, &c.OvnFip.Spec.IPName, false).
		String(constants.FieldKubeOVNOvnFipVpc, &c.OvnFip.Spec.Vpc, false).
		String(constants.FieldKubeOVNOvnFipV4IP, &c.OvnFip.Spec.V4Ip, false).
		String(constants.FieldKubeOVNOvnFipV6IP, &c.OvnFip.Spec.V6Ip, false)
}

func (c *Constructor) Validate() error {
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.OvnFip, nil
}

func Creator(name string) util.Constructor {
	return &Constructor{
		OvnFip: &kubeovnv1.OvnFip{
			ObjectMeta: util.NewObjectMeta("", name),
		},
	}
}

func Updater(obj *kubeovnv1.OvnFip) util.Constructor {
	return &Constructor{OvnFip: obj}
}
