package kubeovn_switch_lb_rule

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var _ util.Constructor = &Constructor{}

type Constructor struct {
	SwitchLBRule *kubeovnv1.SwitchLBRule
}

func parseSlrPort(i interface{}) kubeovnv1.SlrPort {
	m := i.(map[string]interface{})
	return kubeovnv1.SlrPort{
		Name:       m[constants.FieldKubeOVNSlrPortName].(string),
		Port:       int32(m[constants.FieldKubeOVNSlrPortPort].(int)),
		TargetPort: int32(m[constants.FieldKubeOVNSlrPortTargetPort].(int)),
		Protocol:   m[constants.FieldKubeOVNSlrPortProtocol].(string),
	}
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.SwitchLBRule.Labels).
		Labels(&c.SwitchLBRule.Labels).
		Description(&c.SwitchLBRule.Annotations).
		String(constants.FieldKubeOVNSwitchLBRuleVip, &c.SwitchLBRule.Spec.Vip, true).
		String(constants.FieldKubeOVNSwitchLBRuleNamespace, &c.SwitchLBRule.Spec.Namespace, false).
		String(constants.FieldKubeOVNSwitchLBRuleSessionAffinity, &c.SwitchLBRule.Spec.SessionAffinity, false)

	customProcessors := []util.Processor{
		{
			Field: constants.FieldKubeOVNSwitchLBRuleSelector,
			Parser: func(i interface{}) error {
				c.SwitchLBRule.Spec.Selector = append(c.SwitchLBRule.Spec.Selector, i.(string))
				return nil
			},
		},
		{
			Field: constants.FieldKubeOVNSwitchLBRuleEndpoints,
			Parser: func(i interface{}) error {
				c.SwitchLBRule.Spec.Endpoints = append(c.SwitchLBRule.Spec.Endpoints, i.(string))
				return nil
			},
		},
		{
			Field:    constants.FieldKubeOVNSwitchLBRulePorts,
			Required: true,
			Parser: func(i interface{}) error {
				c.SwitchLBRule.Spec.Ports = append(c.SwitchLBRule.Spec.Ports, parseSlrPort(i))
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
	return c.SwitchLBRule, nil
}

func Creator(name string) util.Constructor {
	return &Constructor{
		SwitchLBRule: &kubeovnv1.SwitchLBRule{
			ObjectMeta: util.NewObjectMeta("", name),
		},
	}
}

func Updater(obj *kubeovnv1.SwitchLBRule) util.Constructor {
	obj.Spec.Selector = nil
	obj.Spec.Endpoints = nil
	obj.Spec.Ports = nil
	return &Constructor{SwitchLBRule: obj}
}
