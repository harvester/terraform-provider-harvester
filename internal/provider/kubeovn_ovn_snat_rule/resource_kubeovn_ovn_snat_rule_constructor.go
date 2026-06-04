package kubeovn_ovn_snat_rule

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var _ util.Constructor = &Constructor{}

type Constructor struct {
	OvnSnatRule *kubeovnv1.OvnSnatRule
}

func (c *Constructor) Setup() util.Processors {
	return util.NewProcessors().
		Tags(&c.OvnSnatRule.Labels).
		Labels(&c.OvnSnatRule.Labels).
		Description(&c.OvnSnatRule.Annotations).
		String(constants.FieldKubeOVNOvnSnatOvnEip, &c.OvnSnatRule.Spec.OvnEip, true).
		String(constants.FieldKubeOVNOvnSnatVpcSubnet, &c.OvnSnatRule.Spec.VpcSubnet, true).
		String(constants.FieldKubeOVNOvnSnatIPName, &c.OvnSnatRule.Spec.IPName, false).
		String(constants.FieldKubeOVNOvnSnatVpc, &c.OvnSnatRule.Spec.Vpc, false).
		String(constants.FieldKubeOVNOvnSnatV4IpCidr, &c.OvnSnatRule.Spec.V4IpCidr, false).
		String(constants.FieldKubeOVNOvnSnatV6IpCidr, &c.OvnSnatRule.Spec.V6IpCidr, false)
}

func (c *Constructor) Validate() error {
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.OvnSnatRule, nil
}

func Creator(name string) util.Constructor {
	return &Constructor{
		OvnSnatRule: &kubeovnv1.OvnSnatRule{
			ObjectMeta: util.NewObjectMeta("", name),
		},
	}
}

func Updater(obj *kubeovnv1.OvnSnatRule) util.Constructor {
	return &Constructor{OvnSnatRule: obj}
}
