package kubeovn_switch_lb_rule

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func slrPortSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			constants.FieldKubeOVNSlrPortName: {
				Type:     schema.TypeString,
				Required: true,
			},
			constants.FieldKubeOVNSlrPortPort: {
				Type:     schema.TypeInt,
				Required: true,
			},
			constants.FieldKubeOVNSlrPortTargetPort: {
				Type:     schema.TypeInt,
				Optional: true,
			},
			constants.FieldKubeOVNSlrPortProtocol: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNSwitchLBRuleVip: {
			Type:     schema.TypeString,
			Required: true,
		},
		constants.FieldKubeOVNSwitchLBRuleNamespace: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNSwitchLBRuleSelector: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		constants.FieldKubeOVNSwitchLBRuleEndpoints: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		constants.FieldKubeOVNSwitchLBRuleSessionAffinity: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNSwitchLBRulePorts: {
			Type:     schema.TypeList,
			Required: true,
			Elem:     slrPortSchema(),
		},
		constants.FieldKubeOVNSwitchLBRuleStatusPorts: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNSwitchLBRuleStatusService: {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
