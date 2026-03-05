package kubeovn_vpc

import (
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

var _ util.Constructor = &Constructor{}

type Constructor struct {
	Vpc *kubeovnv1.Vpc
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.Vpc.Labels).
		Labels(&c.Vpc.Labels).
		Description(&c.Vpc.Annotations).
		Bool(constants.FieldKubeOVNVpcEnableExternal, &c.Vpc.Spec.EnableExternal, false).
		Bool(constants.FieldKubeOVNVpcEnableBfd, &c.Vpc.Spec.EnableBfd, false)

	customProcessors := []util.Processor{
		{
			Field: constants.FieldKubeOVNVpcNamespaces,
			Parser: func(i interface{}) error {
				c.Vpc.Spec.Namespaces = append(c.Vpc.Spec.Namespaces, i.(string))
				return nil
			},
		},
		{
			Field: constants.FieldKubeOVNVpcStaticRoutes,
			Parser: func(i interface{}) error {
				route := i.(map[string]interface{})
				c.Vpc.Spec.StaticRoutes = append(c.Vpc.Spec.StaticRoutes, &kubeovnv1.StaticRoute{
					Policy:     kubeovnv1.RoutePolicy(route[constants.FieldKubeOVNStaticRoutePolicy].(string)),
					CIDR:       route[constants.FieldKubeOVNStaticRouteCIDR].(string),
					NextHopIP:  route[constants.FieldKubeOVNStaticRouteNextHopIP].(string),
					ECMPMode:   route[constants.FieldKubeOVNStaticRouteECMPMode].(string),
					RouteTable: route[constants.FieldKubeOVNStaticRouteTable].(string),
				})
				return nil
			},
		},
		{
			Field: constants.FieldKubeOVNVpcPolicyRoutes,
			Parser: func(i interface{}) error {
				route := i.(map[string]interface{})
				c.Vpc.Spec.PolicyRoutes = append(c.Vpc.Spec.PolicyRoutes, &kubeovnv1.PolicyRoute{
					Priority:  route[constants.FieldKubeOVNPolicyRoutePriority].(int),
					Match:     route[constants.FieldKubeOVNPolicyRouteMatch].(string),
					Action:    kubeovnv1.PolicyRouteAction(route[constants.FieldKubeOVNPolicyRouteAction].(string)),
					NextHopIP: route[constants.FieldKubeOVNPolicyRouteNextHopIP].(string),
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
	return c.Vpc, nil
}

func Creator(name string) util.Constructor {
	vpc := &kubeovnv1.Vpc{
		ObjectMeta: util.NewObjectMeta("", name),
	}
	return &Constructor{Vpc: vpc}
}

func Updater(vpc *kubeovnv1.Vpc) util.Constructor {
	vpc.Spec.Namespaces = nil
	vpc.Spec.StaticRoutes = nil
	vpc.Spec.PolicyRoutes = nil
	return &Constructor{Vpc: vpc}
}
