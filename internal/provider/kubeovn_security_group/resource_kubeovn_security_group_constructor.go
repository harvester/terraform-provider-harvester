package kubeovn_security_group

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var _ util.Constructor = &Constructor{}

type Constructor struct {
	SG *kubeovnv1.SecurityGroup
}

func parseSgRule(i interface{}) *kubeovnv1.SgRule {
	r := i.(map[string]interface{})
	return &kubeovnv1.SgRule{
		IPVersion:           r[constants.FieldKubeOVNSGRuleIPVersion].(string),
		Protocol:            kubeovnv1.SgProtocol(r[constants.FieldKubeOVNSGRuleProtocol].(string)),
		Priority:            r[constants.FieldKubeOVNSGRulePriority].(int),
		RemoteType:          kubeovnv1.SgRemoteType(r[constants.FieldKubeOVNSGRuleRemoteType].(string)),
		RemoteAddress:       r[constants.FieldKubeOVNSGRuleRemoteAddress].(string),
		RemoteSecurityGroup: r[constants.FieldKubeOVNSGRuleRemoteSecurityGroup].(string),
		PortRangeMin:        r[constants.FieldKubeOVNSGRulePortRangeMin].(int),
		PortRangeMax:        r[constants.FieldKubeOVNSGRulePortRangeMax].(int),
		Policy:              kubeovnv1.SgPolicy(r[constants.FieldKubeOVNSGRulePolicy].(string)),
	}
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.SG.Labels).
		Labels(&c.SG.Labels).
		Description(&c.SG.Annotations).
		Bool(constants.FieldKubeOVNSGAllowSameGroupTraffic, &c.SG.Spec.AllowSameGroupTraffic, false)

	customProcessors := []util.Processor{
		{
			Field: constants.FieldKubeOVNSGIngressRules,
			Parser: func(i interface{}) error {
				c.SG.Spec.IngressRules = append(c.SG.Spec.IngressRules, parseSgRule(i))
				return nil
			},
		},
		{
			Field: constants.FieldKubeOVNSGEgressRules,
			Parser: func(i interface{}) error {
				c.SG.Spec.EgressRules = append(c.SG.Spec.EgressRules, parseSgRule(i))
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
	return c.SG, nil
}

func Creator(name string) util.Constructor {
	return &Constructor{
		SG: &kubeovnv1.SecurityGroup{
			ObjectMeta: util.NewObjectMeta("", name),
		},
	}
}

func Updater(obj *kubeovnv1.SecurityGroup) util.Constructor {
	obj.Spec.IngressRules = nil
	obj.Spec.EgressRules = nil
	return &Constructor{SG: obj}
}
