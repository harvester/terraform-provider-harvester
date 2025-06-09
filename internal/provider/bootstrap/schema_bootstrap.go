package bootstrap

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldBootstrapAPIURL: {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.NoZeroValues,
			Description:  "API URL in the harvester",
		},
		constants.FieldBootstrapInitialPassword: {
			Type:         schema.TypeString,
			Optional:     true,
			Sensitive:    true,
			ForceNew:     true,
			Default:      "admin",
			ValidateFunc: validation.NoZeroValues,
			Description:  "Default password in the harvester",
		},
		constants.FieldBootstrapPassword: {
			Type:         schema.TypeString,
			Required:     true,
			Sensitive:    true,
			ForceNew:     true,
			ValidateFunc: validation.NoZeroValues,
			Description:  "New password for admin user",
		},
		constants.FieldBootstrapKubeConfig: {
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			Default:      "~/.kube/config",
			ValidateFunc: validation.NoZeroValues,
			Description:  "Path to store the kubeconfig file",
		},
		constants.FieldShouldUpdatePassword: {
			Type:     schema.TypeBool,
			Computed: true,
		},
	}
	return s
}
