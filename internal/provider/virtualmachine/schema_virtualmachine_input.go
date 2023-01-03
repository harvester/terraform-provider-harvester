package virtualmachine

import (
	"github.com/harvester/harvester/pkg/builder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func resourceInputSchema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldInputName: {
			Type:     schema.TypeString,
			Required: true,
		},
		constants.FieldInputType: {
			Type:     schema.TypeString,
			Optional: true,
			Default:  builder.InputTypeTablet,
			ValidateFunc: validation.StringInSlice([]string{
				builder.InputTypeTablet,
			}, false),
		},
		constants.FieldInputBus: {
			Type:     schema.TypeString,
			Optional: true,
			Default:  builder.InputBusUSB,
			ValidateFunc: validation.StringInSlice([]string{
				builder.InputBusUSB,
				builder.InputBusVirtio,
			}, false),
		},
	}
	return s
}
