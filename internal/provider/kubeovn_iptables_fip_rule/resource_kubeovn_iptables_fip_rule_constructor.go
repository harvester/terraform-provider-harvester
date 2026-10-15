package kubeovn_iptables_fip_rule

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var _ util.Constructor = &Constructor{}

type Constructor struct {
	FIPRule *kubeovnv1.IptablesFIPRule
}

func (c *Constructor) Setup() util.Processors {
	return util.NewProcessors().
		Tags(&c.FIPRule.Labels).
		Labels(&c.FIPRule.Labels).
		Description(&c.FIPRule.Annotations).
		String(constants.FieldKubeOVNIptablesFIPEIP, &c.FIPRule.Spec.EIP, true).
		String(constants.FieldKubeOVNIptablesFIPInternalIP, &c.FIPRule.Spec.InternalIP, true)
}

func (c *Constructor) Validate() error {
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.FIPRule, nil
}

func Creator(name string) util.Constructor {
	return &Constructor{
		FIPRule: &kubeovnv1.IptablesFIPRule{
			ObjectMeta: util.NewObjectMeta("", name),
		},
	}
}

func Updater(obj *kubeovnv1.IptablesFIPRule) util.Constructor {
	return &Constructor{FIPRule: obj}
}
