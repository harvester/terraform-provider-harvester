package virtualmachine

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func resourceAccessCredentialSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		constants.FieldAccessCredentialSSHPublicKey: {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldAccessCredentialSecretName: {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Name of the Kubernetes secret containing SSH public keys",
					},
					constants.FieldAccessCredentialPropagationMethod: {
						Type:     schema.TypeString,
						Required: true,
						ValidateFunc: validation.StringInSlice([]string{
							"configDrive",
							"noCloud",
							"qemuGuestAgent",
						}, false),
						Description: "Method to propagate SSH keys: configDrive, noCloud, or qemuGuestAgent",
					},
					constants.FieldAccessCredentialUsers: {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Description: "List of guest users for qemuGuestAgent propagation (required when propagation_method is qemuGuestAgent)",
					},
				},
			},
			Description: "SSH public key access credential sourced from a Kubernetes secret",
		},
		constants.FieldAccessCredentialUserPassword: {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldAccessCredentialSecretName: {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Name of the Kubernetes secret containing user passwords",
					},
				},
			},
			Description: "User password access credential sourced from a Kubernetes secret, propagated via qemu guest agent",
		},
	}
}
