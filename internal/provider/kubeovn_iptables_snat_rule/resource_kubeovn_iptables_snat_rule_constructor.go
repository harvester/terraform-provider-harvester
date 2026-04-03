package kubeovn_iptables_snat_rule

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var _ util.Constructor = &Constructor{}

type Constructor struct {
	SnatRule *kubeovnv1.IptablesSnatRule
}

func (c *Constructor) Setup() util.Processors {
	return util.NewProcessors().
		Tags(&c.SnatRule.Labels).
		Labels(&c.SnatRule.Labels).
		Description(&c.SnatRule.Annotations).
		String(constants.FieldKubeOVNIptablesSnatEIP, &c.SnatRule.Spec.EIP, true).
		String(constants.FieldKubeOVNIptablesSnatInternalCIDR, &c.SnatRule.Spec.InternalCIDR, true)
}

func (c *Constructor) Validate() error {
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.SnatRule, nil
}

func Creator(name string) util.Constructor {
	return &Constructor{
		SnatRule: &kubeovnv1.IptablesSnatRule{
			ObjectMeta: util.NewObjectMeta("", name),
		},
	}
}

func Updater(obj *kubeovnv1.IptablesSnatRule) util.Constructor {
	return &Constructor{SnatRule: obj}
}
