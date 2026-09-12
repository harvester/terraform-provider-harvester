package kubeovn_qos_policy

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var _ util.Constructor = &Constructor{}

type Constructor struct {
	QoS *kubeovnv1.QoSPolicy
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.QoS.Labels).
		Labels(&c.QoS.Labels).
		Description(&c.QoS.Annotations).
		Bool(constants.FieldKubeOVNQoSShared, &c.QoS.Spec.Shared, false)

	customProcessors := []util.Processor{
		{
			Field: constants.FieldKubeOVNQoSBindingType,
			Parser: func(i interface{}) error {
				c.QoS.Spec.BindingType = kubeovnv1.QoSPolicyBindingType(i.(string))
				return nil
			},
		},
		{
			Field: constants.FieldKubeOVNQoSBandwidthLimitRules,
			Parser: func(i interface{}) error {
				r := i.(map[string]interface{})
				c.QoS.Spec.BandwidthLimitRules = append(c.QoS.Spec.BandwidthLimitRules, kubeovnv1.QoSPolicyBandwidthLimitRule{
					Name:       r[constants.FieldKubeOVNQoSRuleName].(string),
					Interface:  r[constants.FieldKubeOVNQoSRuleInterface].(string),
					RateMax:    r[constants.FieldKubeOVNQoSRuleRateMax].(string),
					BurstMax:   r[constants.FieldKubeOVNQoSRuleBurstMax].(string),
					Priority:   r[constants.FieldKubeOVNQoSRulePriority].(int),
					Direction:  kubeovnv1.QoSPolicyRuleDirection(r[constants.FieldKubeOVNQoSRuleDirection].(string)),
					MatchType:  kubeovnv1.QoSPolicyRuleMatchType(r[constants.FieldKubeOVNQoSRuleMatchType].(string)),
					MatchValue: r[constants.FieldKubeOVNQoSRuleMatchValue].(string),
				})
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
	return c.QoS, nil
}

func Creator(name string) util.Constructor {
	return &Constructor{
		QoS: &kubeovnv1.QoSPolicy{
			ObjectMeta: util.NewObjectMeta("", name),
		},
	}
}

func Updater(obj *kubeovnv1.QoSPolicy) util.Constructor {
	obj.Spec.BandwidthLimitRules = nil
	return &Constructor{QoS: obj}
}
