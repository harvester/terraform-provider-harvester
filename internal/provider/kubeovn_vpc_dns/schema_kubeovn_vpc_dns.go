package kubeovn_vpc_dns

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNVpcDnsReplicas: {
			Type:     schema.TypeInt,
			Required: true,
		},
		constants.FieldKubeOVNVpcDnsVpc: {
			Type:     schema.TypeString,
			Required: true,
		},
		constants.FieldKubeOVNVpcDnsSubnet: {
			Type:     schema.TypeString,
			Required: true,
		},
		constants.FieldKubeOVNVpcDnsStatusActive: {
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
