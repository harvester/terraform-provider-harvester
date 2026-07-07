package kubeovn_iptables_eip

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var _ util.Constructor = &Constructor{}

type Constructor struct {
	EIP *kubeovnv1.IptablesEIP
}

func (c *Constructor) Setup() util.Processors {
	return util.NewProcessors().
		Tags(&c.EIP.Labels).
		Labels(&c.EIP.Labels).
		Description(&c.EIP.Annotations).
		String(constants.FieldKubeOVNIptablesEIPV4IP, &c.EIP.Spec.V4ip, false).
		String(constants.FieldKubeOVNIptablesEIPV6IP, &c.EIP.Spec.V6ip, false).
		String(constants.FieldKubeOVNIptablesEIPMacAddress, &c.EIP.Spec.MacAddress, false).
		String(constants.FieldKubeOVNIptablesEIPNatGwDp, &c.EIP.Spec.NatGwDp, true).
		String(constants.FieldKubeOVNIptablesEIPQoSPolicy, &c.EIP.Spec.QoSPolicy, false).
		String(constants.FieldKubeOVNIptablesEIPExternalSubnet, &c.EIP.Spec.ExternalSubnet, false)
}

func (c *Constructor) Validate() error {
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.EIP, nil
}

func Creator(name string) util.Constructor {
	return &Constructor{
		EIP: &kubeovnv1.IptablesEIP{
			ObjectMeta: util.NewObjectMeta("", name),
		},
	}
}

func Updater(obj *kubeovnv1.IptablesEIP) util.Constructor {
	return &Constructor{EIP: obj}
}
