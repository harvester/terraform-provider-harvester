package cloudinitsecret

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldCloudInitSecretUserData: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldCloudInitSecretUserDataBase64: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldCloudInitSecretNetworkData: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldCloudInitSecretNetworkDataBase64: {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
	util.NamespacedSchemaWrap(s, false)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
