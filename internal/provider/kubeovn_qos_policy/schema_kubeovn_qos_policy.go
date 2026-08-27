package kubeovn_qos_policy

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func bandwidthLimitRuleSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			constants.FieldKubeOVNQoSRuleName: {
				Type:     schema.TypeString,
				Required: true,
			},
			constants.FieldKubeOVNQoSRuleInterface: {
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.FieldKubeOVNQoSRuleRateMax: {
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.FieldKubeOVNQoSRuleBurstMax: {
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.FieldKubeOVNQoSRulePriority: {
				Type:     schema.TypeInt,
				Optional: true,
			},
			constants.FieldKubeOVNQoSRuleDirection: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"ingress", "egress"}, false),
			},
			constants.FieldKubeOVNQoSRuleMatchType: {
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.FieldKubeOVNQoSRuleMatchValue: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNQoSShared: {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		constants.FieldKubeOVNQoSBindingType: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"EIP", "NATGW"}, false),
		},
		constants.FieldKubeOVNQoSBandwidthLimitRules: {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     bandwidthLimitRuleSchema(),
		},
		constants.FieldKubeOVNQoSStatusShared: {
			Type:     schema.TypeBool,
			Computed: true,
		},
		constants.FieldKubeOVNQoSStatusBindingType: {
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
