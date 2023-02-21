package virtualmachine

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func resourceTPMSchema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldTPMName: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "just add this field for doc generation",
		},
	}
	return s
}
