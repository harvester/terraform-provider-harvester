package kubeovn_ovn_eip

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var _ util.Constructor = &Constructor{}

type Constructor struct {
	OvnEip *kubeovnv1.OvnEip
}

func (c *Constructor) Setup() util.Processors {
	return util.NewProcessors().
		Tags(&c.OvnEip.Labels).
		Labels(&c.OvnEip.Labels).
		Description(&c.OvnEip.Annotations).
		String(constants.FieldKubeOVNOvnEipExternalSubnet, &c.OvnEip.Spec.ExternalSubnet, true).
		String(constants.FieldKubeOVNOvnEipV4IP, &c.OvnEip.Spec.V4Ip, false).
		String(constants.FieldKubeOVNOvnEipV6IP, &c.OvnEip.Spec.V6Ip, false).
		String(constants.FieldKubeOVNOvnEipMacAddress, &c.OvnEip.Spec.MacAddress, false)
}

func (c *Constructor) Validate() error {
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.OvnEip, nil
}

func Creator(name string) util.Constructor {
	return &Constructor{
		OvnEip: &kubeovnv1.OvnEip{
			ObjectMeta: util.NewObjectMeta("", name),
		},
	}
}

func Updater(obj *kubeovnv1.OvnEip) util.Constructor {
	return &Constructor{OvnEip: obj}
}
