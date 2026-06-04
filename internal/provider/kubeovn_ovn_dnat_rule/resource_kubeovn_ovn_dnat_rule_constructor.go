package kubeovn_ovn_dnat_rule

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var _ util.Constructor = &Constructor{}

type Constructor struct {
	OvnDnatRule *kubeovnv1.OvnDnatRule
}

func (c *Constructor) Setup() util.Processors {
	return util.NewProcessors().
		Tags(&c.OvnDnatRule.Labels).
		Labels(&c.OvnDnatRule.Labels).
		Description(&c.OvnDnatRule.Annotations).
		String(constants.FieldKubeOVNOvnDnatOvnEip, &c.OvnDnatRule.Spec.OvnEip, true).
		String(constants.FieldKubeOVNOvnDnatIPType, &c.OvnDnatRule.Spec.IPType, false).
		String(constants.FieldKubeOVNOvnDnatIPName, &c.OvnDnatRule.Spec.IPName, false).
		String(constants.FieldKubeOVNOvnDnatInternalPort, &c.OvnDnatRule.Spec.InternalPort, false).
		String(constants.FieldKubeOVNOvnDnatExternalPort, &c.OvnDnatRule.Spec.ExternalPort, false).
		String(constants.FieldKubeOVNOvnDnatProtocol, &c.OvnDnatRule.Spec.Protocol, false).
		String(constants.FieldKubeOVNOvnDnatVpc, &c.OvnDnatRule.Spec.Vpc, false).
		String(constants.FieldKubeOVNOvnDnatV4IP, &c.OvnDnatRule.Spec.V4Ip, false).
		String(constants.FieldKubeOVNOvnDnatV6IP, &c.OvnDnatRule.Spec.V6Ip, false)
}

func (c *Constructor) Validate() error {
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.OvnDnatRule, nil
}

func Creator(name string) util.Constructor {
	return &Constructor{
		OvnDnatRule: &kubeovnv1.OvnDnatRule{
			ObjectMeta: util.NewObjectMeta("", name),
		},
	}
}

func Updater(obj *kubeovnv1.OvnDnatRule) util.Constructor {
	return &Constructor{OvnDnatRule: obj}
}
