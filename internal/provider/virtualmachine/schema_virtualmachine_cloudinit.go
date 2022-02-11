package virtualmachine

import (
	"github.com/harvester/harvester/pkg/builder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func resourceCloudInitSchema() map[string]*schema.Schema {
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
			Type:          schema.TypeString,
			ConflictsWith: []string{constants.FieldCloudInitNetworkDataBase64, constants.FieldCloudInitNetworkDataSecretName},
			Optional:      true,
		},
		constants.FieldCloudInitNetworkDataBase64: {
			Type:          schema.TypeString,
			ConflictsWith: []string{constants.FieldCloudInitNetworkData, constants.FieldCloudInitNetworkDataSecretName},
			Optional:      true,
		},
		constants.FieldCloudInitNetworkDataSecretName: {
			Type:          schema.TypeString,
			ConflictsWith: []string{constants.FieldCloudInitNetworkData, constants.FieldCloudInitNetworkDataBase64},
			Optional:      true,
		},
		constants.FieldCloudInitUserData: {
			Type:          schema.TypeString,
			ConflictsWith: []string{constants.FieldCloudInitUserDataBase64, constants.FieldCloudInitUserDataSecretName},
			Optional:      true,
		},
		constants.FieldCloudInitUserDataBase64: {
			Type:          schema.TypeString,
			ConflictsWith: []string{constants.FieldCloudInitUserData, constants.FieldCloudInitUserDataSecretName},
			Optional:      true,
		},
		constants.FieldCloudInitUserDataSecretName: {
			Type:          schema.TypeString,
			ConflictsWith: []string{constants.FieldCloudInitUserData, constants.FieldCloudInitUserDataBase64},
			Optional:      true,
		},
	}
	return s
}
