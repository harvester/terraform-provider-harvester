package virtualmachine

import (
	"github.com/harvester/harvester/pkg/builder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func cloudInitSchema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldCloudInitType: {
			Type:     schema.TypeString,
			Optional: true,
			Default:  builder.CloudInitTypeNoCloud,
			ValidateFunc: validation.StringInSlice([]string{
				builder.CloudInitTypeNoCloud,
				builder.CloudInitTypeConfigDrive,
			}, false),
		},
		constants.FieldCloudInitNetworkData: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldCloudInitNetworkDataBase64: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldCloudInitNetworkDataSecretName: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldCloudInitUserData: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldCloudInitUserDataBase64: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldCloudInitUserDataSecretName: {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
	return s
}
