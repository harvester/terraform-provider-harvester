package virtualmachine

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func resourceClockSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		constants.FieldClockTimezone: {
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{constants.FieldVirtualMachineClock + ".0." + constants.FieldClockUTCOffsetSeconds},
			Description:   "Timezone for the guest clock (e.g. 'America/New_York'). Mutually exclusive with utc_offset_seconds",
		},
		constants.FieldClockUTCOffsetSeconds: {
			Type:          schema.TypeInt,
			Optional:      true,
			ConflictsWith: []string{constants.FieldVirtualMachineClock + ".0." + constants.FieldClockTimezone},
			Description:   "UTC offset in seconds. Mutually exclusive with timezone",
		},
		constants.FieldClockTimer: {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: resourceClockTimerSchema(),
			},
			Description: "Timer configuration for the guest clock",
		},
	}
}

func resourceClockTimerSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		constants.FieldTimerHPET: {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldTimerEnabled: {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  true,
					},
					constants.FieldTimerTickPolicy: {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringInSlice([]string{"delay", "catchup", "merge", "discard"}, false),
					},
				},
			},
		},
		constants.FieldTimerKVM: {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldTimerEnabled: {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  true,
					},
				},
			},
		},
		constants.FieldTimerPIT: {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldTimerEnabled: {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  true,
					},
					constants.FieldTimerTickPolicy: {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringInSlice([]string{"delay", "catchup", "discard"}, false),
					},
				},
			},
		},
		constants.FieldTimerRTC: {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldTimerEnabled: {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  true,
					},
					constants.FieldTimerTickPolicy: {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringInSlice([]string{"delay", "catchup"}, false),
					},
					constants.FieldTimerTrack: {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringInSlice([]string{"guest", "wall"}, false),
					},
				},
			},
		},
		constants.FieldTimerHyperv: {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldTimerEnabled: {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  true,
					},
				},
			},
		},
	}
}
