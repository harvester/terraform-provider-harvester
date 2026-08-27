package kubeovn_iptables_dnat_rule

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var _ util.Constructor = &Constructor{}

type Constructor struct {
	DnatRule *kubeovnv1.IptablesDnatRule
}

func (c *Constructor) Setup() util.Processors {
	return util.NewProcessors().
		Tags(&c.DnatRule.Labels).
		Labels(&c.DnatRule.Labels).
		Description(&c.DnatRule.Annotations).
		String(constants.FieldKubeOVNIptablesDnatEIP, &c.DnatRule.Spec.EIP, true).
		String(constants.FieldKubeOVNIptablesDnatExternalPort, &c.DnatRule.Spec.ExternalPort, true).
		String(constants.FieldKubeOVNIptablesDnatProtocol, &c.DnatRule.Spec.Protocol, false).
		String(constants.FieldKubeOVNIptablesDnatInternalIP, &c.DnatRule.Spec.InternalIP, true).
		String(constants.FieldKubeOVNIptablesDnatInternalPort, &c.DnatRule.Spec.InternalPort, true)
}

func (c *Constructor) Validate() error {
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.DnatRule, nil
}

func Creator(name string) util.Constructor {
	return &Constructor{
		DnatRule: &kubeovnv1.IptablesDnatRule{
			ObjectMeta: util.NewObjectMeta("", name),
		},
	}
}

func Updater(obj *kubeovnv1.IptablesDnatRule) util.Constructor {
	return &Constructor{DnatRule: obj}
}
