package kubeovn_ovn_eip

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNOvnEipExternalSubnet: {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		constants.FieldKubeOVNOvnEipV4IP: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		constants.FieldKubeOVNOvnEipV6IP: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		constants.FieldKubeOVNOvnEipMacAddress: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		constants.FieldKubeOVNOvnEipType: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldKubeOVNOvnEipStatusReady: {
			Type:     schema.TypeBool,
			Computed: true,
		},
		constants.FieldKubeOVNOvnEipStatusV4IP: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNOvnEipStatusV6IP: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNOvnEipStatusMac: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNOvnEipStatusNat: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldKubeOVNOvnEipStatusType: {
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
