package virtualmachine

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func resourceHypervSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		constants.FieldHypervRelaxed: {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Relaxed instructs the guest OS to disable watchdog timeouts",
		},
		constants.FieldHypervVAPIC: {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "VAPIC improves the paravirtualized handling of interrupts",
		},
		constants.FieldHypervVPIndex: {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "VPIndex enables the Virtual Processor Index to help Windows identifying virtual processors",
		},
		constants.FieldHypervRuntime: {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Runtime improves the time accounting to improve scheduling in the guest",
		},
		constants.FieldHypervSyNIC: {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "SyNIC enables the Synthetic Interrupt Controller",
		},
		constants.FieldHypervReset: {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Reset enables Hyper-V reboot/reset for the VM. Requires synic",
		},
		constants.FieldHypervFrequencies: {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Frequencies improves the TSC clock source handling for Hyper-V on KVM",
		},
		constants.FieldHypervReenlightenment: {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Reenlightenment enables the notifications on TSC frequency changes",
		},
		constants.FieldHypervTLBFlush: {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "TLBFlush improves performance in overcommitted environments. Requires vpindex",
		},
		constants.FieldHypervIPI: {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "IPI improves performance in overcommitted environments. Requires vpindex",
		},
		constants.FieldHypervEVMCS: {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "EVMCS speeds up L2 vmexits, but disables other virtualization features. Requires vapic",
		},
		constants.FieldHypervSpinlocks: {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Spinlocks enables the spinlock retry mechanism",
		},
		constants.FieldHypervSpinlocksRetries: {
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          4096,
			ValidateFunc:     validation.IntAtLeast(4096),
			Description:      "Number of spinlock retries. Must be >= 4096. Only used when spinlocks is true",
			DiffSuppressFunc: suppressSpinlocksRetriesIfDisabled,
		},
		constants.FieldHypervSyNICTimer: {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "SyNICTimer enables Synthetic Interrupt Controller Timers, reducing CPU load",
		},
		constants.FieldHypervSyNICTimerDirect: {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "SyNICTimer direct mode. Only used when synictimer is true",
		},
		constants.FieldHypervVendorID: {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "VendorID allows setting the hypervisor vendor ID",
		},
		constants.FieldHypervVendorIDValue: {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateFunc:     validation.StringLenBetween(0, 12),
			Description:      "Hypervisor vendor ID string, up to 12 characters. Only used when vendorid is true",
			DiffSuppressFunc: suppressVendorIDValueIfDisabled,
		},
	}
}

func suppressSpinlocksRetriesIfDisabled(_, _, _ string, d *schema.ResourceData) bool {
	hypervList := d.Get(constants.FieldVirtualMachineHyperv).([]interface{})
	if len(hypervList) == 0 {
		return true
	}
	hv := hypervList[0].(map[string]interface{})
	return !hv[constants.FieldHypervSpinlocks].(bool)
}

func suppressVendorIDValueIfDisabled(_, _, _ string, d *schema.ResourceData) bool {
	hypervList := d.Get(constants.FieldVirtualMachineHyperv).([]interface{})
	if len(hypervList) == 0 {
		return true
	}
	hv := hypervList[0].(map[string]interface{})
	return !hv[constants.FieldHypervVendorID].(bool)
}
