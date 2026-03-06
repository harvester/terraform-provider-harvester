package kubeovn_vip

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNVipNamespace: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNVipSubnet: {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		constants.FieldKubeOVNVipType: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNVipV4IP: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		constants.FieldKubeOVNVipV6IP: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		constants.FieldKubeOVNVipMacAddress: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		constants.FieldKubeOVNVipSelector: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		constants.FieldKubeOVNVipAttachSubnets: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		constants.FieldKubeOVNVipStatusV4IP: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNVipStatusV6IP: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNVipStatusMac: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNVipStatusType: {
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
