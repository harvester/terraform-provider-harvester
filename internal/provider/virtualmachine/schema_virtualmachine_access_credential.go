package virtualmachine

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/harvester/pkg/builder"

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
							builder.CloudInitTypeConfigDrive,
							builder.CloudInitTypeNoCloud,
							constants.AccessCredentialPropagationQemuGuestAgent,
						}, false),
						Description: "Method to propagate SSH keys: configDrive, noCloud, or qemuGuestAgent",
					},
					constants.FieldAccessCredentialUsers: {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Description: "List of guest users to receive the SSH keys, only for qemuGuestAgent propagation (required with qemuGuestAgent, rejected with configDrive and noCloud)",
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
