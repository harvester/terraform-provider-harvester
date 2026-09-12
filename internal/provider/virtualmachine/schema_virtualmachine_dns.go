package virtualmachine

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func resourceDNSConfigSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		constants.FieldDNSConfigNameservers: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "List of DNS nameservers",
		},
		constants.FieldDNSConfigSearches: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "List of DNS search domains",
		},
		constants.FieldDNSConfigOptions: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldDNSOptionName: {
						Type:        schema.TypeString,
						Required:    true,
						Description: "DNS option name",
					},
					constants.FieldDNSOptionValue: {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "DNS option value",
					},
				},
			},
			Description: "List of DNS resolver options",
		},
	}
}
