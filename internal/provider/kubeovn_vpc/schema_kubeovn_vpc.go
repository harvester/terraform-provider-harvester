package kubeovn_vpc

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNVpcNamespaces: {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		constants.FieldKubeOVNVpcStaticRoutes: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldKubeOVNStaticRoutePolicy: {
						Type:         schema.TypeString,
						Optional:     true,
						Default:      "",
						ValidateFunc: validation.StringInSlice([]string{"", "policySrc"}, false),
					},
					constants.FieldKubeOVNStaticRouteCIDR: {
						Type:     schema.TypeString,
						Required: true,
					},
					constants.FieldKubeOVNStaticRouteNextHopIP: {
						Type:     schema.TypeString,
						Required: true,
					},
					constants.FieldKubeOVNStaticRouteECMPMode: {
						Type:     schema.TypeString,
						Optional: true,
						Default:  "",
					},
					constants.FieldKubeOVNStaticRouteTable: {
						Type:     schema.TypeString,
						Optional: true,
						Default:  "",
					},
				},
			},
		},
		constants.FieldKubeOVNVpcPolicyRoutes: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldKubeOVNPolicyRoutePriority: {
						Type:     schema.TypeInt,
						Required: true,
					},
					constants.FieldKubeOVNPolicyRouteMatch: {
						Type:     schema.TypeString,
						Required: true,
					},
					constants.FieldKubeOVNPolicyRouteAction: {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice([]string{"allow", "drop", "reroute"}, false),
					},
					constants.FieldKubeOVNPolicyRouteNextHopIP: {
						Type:     schema.TypeString,
						Optional: true,
						Default:  "",
					},
				},
			},
		},
		constants.FieldKubeOVNVpcEnableExternal: {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		constants.FieldKubeOVNVpcEnableBfd: {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		constants.FieldKubeOVNVpcDefaultSubnet: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNVpcStandby: {
			Type:     schema.TypeBool,
			Computed: true,
		},
		constants.FieldKubeOVNVpcRouter: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNVpcSubnets: {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
