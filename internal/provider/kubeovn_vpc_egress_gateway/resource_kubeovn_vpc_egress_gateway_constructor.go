package kubeovn_vpc_egress_gateway

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var _ util.Constructor = &Constructor{}

type Constructor struct {
	EGW *kubeovnv1.VpcEgressGateway
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.EGW.Labels).
		Labels(&c.EGW.Labels).
		Description(&c.EGW.Annotations).
		String(constants.FieldKubeOVNVpcEgressGatewayVpc, &c.EGW.Spec.VPC, false).
		String(constants.FieldKubeOVNVpcEgressGatewayPrefix, &c.EGW.Spec.Prefix, false).
		String(constants.FieldKubeOVNVpcEgressGatewayImage, &c.EGW.Spec.Image, false).
		String(constants.FieldKubeOVNVpcEgressGatewayInternalSubnet, &c.EGW.Spec.InternalSubnet, false).
		String(constants.FieldKubeOVNVpcEgressGatewayExternalSubnet, &c.EGW.Spec.ExternalSubnet, true).
		String(constants.FieldKubeOVNVpcEgressGatewayTrafficPolicy, &c.EGW.Spec.TrafficPolicy, false)

	customProcessors := []util.Processor{
		{
			Field: constants.FieldKubeOVNVpcEgressGatewayReplicas,
			Parser: func(i interface{}) error {
				c.EGW.Spec.Replicas = int32(i.(int))
				return nil
			},
		},
		{
			Field: constants.FieldKubeOVNVpcEgressGatewayInternalIPs,
			Parser: func(i interface{}) error {
				c.EGW.Spec.InternalIPs = append(c.EGW.Spec.InternalIPs, i.(string))
				return nil
			},
		},
		{
			Field: constants.FieldKubeOVNVpcEgressGatewayExternalIPs,
			Parser: func(i interface{}) error {
				c.EGW.Spec.ExternalIPs = append(c.EGW.Spec.ExternalIPs, i.(string))
				return nil
			},
		},
		{
			Field: constants.FieldKubeOVNVpcEgressGatewayBFD,
			Parser: func(i interface{}) error {
				m := i.(map[string]interface{})
				c.EGW.Spec.BFD = kubeovnv1.VpcEgressGatewayBFDConfig{
					Enabled:    m[constants.FieldKubeOVNVpcEgressGatewayBFDEnabled].(bool),
					MinRX:      int32(m[constants.FieldKubeOVNVpcEgressGatewayBFDMinRX].(int)),
					MinTX:      int32(m[constants.FieldKubeOVNVpcEgressGatewayBFDMinTX].(int)),
					Multiplier: int32(m[constants.FieldKubeOVNVpcEgressGatewayBFDMultiplier].(int)),
				}
				return nil
			},
		},
		{
			Field: constants.FieldKubeOVNVpcEgressGatewayPolicies,
			Parser: func(i interface{}) error {
				m := i.(map[string]interface{})
				policy := kubeovnv1.VpcEgressGatewayPolicy{
					SNAT: m[constants.FieldKubeOVNVpcEgressGatewayPolicySNAT].(bool),
				}
				if v, ok := m[constants.FieldKubeOVNVpcEgressGatewayPolicyIPBlocks]; ok {
					for _, item := range v.([]interface{}) {
						policy.IPBlocks = append(policy.IPBlocks, item.(string))
					}
				}
				if v, ok := m[constants.FieldKubeOVNVpcEgressGatewayPolicySubnets]; ok {
					for _, item := range v.([]interface{}) {
						policy.Subnets = append(policy.Subnets, item.(string))
					}
				}
				c.EGW.Spec.Policies = append(c.EGW.Spec.Policies, policy)
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
	return c.EGW, nil
}

func Creator(namespace, name string) util.Constructor {
	return &Constructor{
		EGW: &kubeovnv1.VpcEgressGateway{
			ObjectMeta: util.NewObjectMeta(namespace, name),
		},
	}
}

func Updater(obj *kubeovnv1.VpcEgressGateway) util.Constructor {
	obj.Spec.InternalIPs = nil
	obj.Spec.ExternalIPs = nil
	obj.Spec.Policies = nil
	return &Constructor{EGW: obj}
}
