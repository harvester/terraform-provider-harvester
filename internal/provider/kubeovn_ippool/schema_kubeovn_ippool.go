package kubeovn_ippool

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNIPPoolSubnet: {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		constants.FieldKubeOVNIPPoolIPs: {
			Type:     schema.TypeList,
			Required: true,
			MinItems: 1,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		constants.FieldKubeOVNIPPoolNamespaces: {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		constants.FieldKubeOVNIPPoolV4AvailableIPs: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIPPoolV4AvailableRange: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIPPoolV4UsingIPs: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIPPoolV4UsingRange: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIPPoolV6AvailableIPs: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIPPoolV6AvailableRange: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIPPoolV6UsingIPs: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNIPPoolV6UsingRange: {
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
