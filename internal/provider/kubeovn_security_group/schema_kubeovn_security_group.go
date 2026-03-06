package kubeovn_security_group

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func sgRuleSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			constants.FieldKubeOVNSGRuleIPVersion: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"ipv4", "ipv6"}, false),
			},
			constants.FieldKubeOVNSGRuleProtocol: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "all",
				ValidateFunc: validation.StringInSlice([]string{"all", "icmp", "tcp", "udp"}, false),
			},
			constants.FieldKubeOVNSGRulePriority: {
				Type:     schema.TypeInt,
				Optional: true,
			},
			constants.FieldKubeOVNSGRuleRemoteType: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"address", "securityGroup"}, false),
			},
			constants.FieldKubeOVNSGRuleRemoteAddress: {
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.FieldKubeOVNSGRuleRemoteSecurityGroup: {
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.FieldKubeOVNSGRulePortRangeMin: {
				Type:     schema.TypeInt,
				Optional: true,
			},
			constants.FieldKubeOVNSGRulePortRangeMax: {
				Type:     schema.TypeInt,
				Optional: true,
			},
			constants.FieldKubeOVNSGRulePolicy: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"allow", "drop"}, false),
			},
		},
	}
}

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNSGAllowSameGroupTraffic: {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		constants.FieldKubeOVNSGIngressRules: {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     sgRuleSchema(),
		},
		constants.FieldKubeOVNSGEgressRules: {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     sgRuleSchema(),
		},
		constants.FieldKubeOVNSGStatusPortGroup: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNSGStatusIngressMD5: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNSGStatusEgressMD5: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNSGStatusIngressSynced: {
			Type:     schema.TypeBool,
			Computed: true,
		},
		constants.FieldKubeOVNSGStatusEgressSynced: {
			Type:     schema.TypeBool,
			Computed: true,
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
