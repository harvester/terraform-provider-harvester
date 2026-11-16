package virtualmachine

import (
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

// deleteProtectionFromConfig returns the delete_protection value written in
// the configuration, or nil when it is not set there. d.GetOk cannot tell an
// explicit false from an unset field, and the field is Computed, so the raw
// configuration is the only reliable source.
func deleteProtectionFromConfig(d *schema.ResourceData) *bool {
	return deleteProtectionFromRawConfig(d.GetRawConfig())
}

func deleteProtectionFromRawConfig(raw cty.Value) *bool {
	if raw.IsNull() || !raw.IsKnown() || !raw.Type().IsObjectType() ||
		!raw.Type().HasAttribute(constants.FieldVirtualMachineDeleteProtection) {
		return nil
	}
	value := raw.GetAttr(constants.FieldVirtualMachineDeleteProtection)
	if value.IsNull() || !value.IsKnown() || !value.Type().Equals(cty.Bool) {
		return nil
	}
	enabled := value.True()
	return &enabled
}

// setDeleteProtectionLabel writes the delete protection label when enabled is
// set. When it is nil the VM keeps whatever value it already has: on create
// that lets Harvester apply its default-vm-delete-protection setting.
func setDeleteProtectionLabel(vm *kubevirtv1.VirtualMachine, enabled *bool) {
	if vm == nil || enabled == nil {
		return
	}
	if vm.Labels == nil {
		vm.Labels = map[string]string{}
	}
	vm.Labels[constants.LabelVMDeleteProtection] = fmt.Sprintf("%t", *enabled)
}

// deleteProtectionError explains why Harvester refuses to delete the VM.
func deleteProtectionError(vm *kubevirtv1.VirtualMachine) error {
	if vm == nil || !importer.IsVMDeleteProtected(vm.Labels) {
		return nil
	}
	return fmt.Errorf("virtual machine %s/%s has delete protection enabled: set %s = false and apply before destroying it",
		vm.Namespace, vm.Name, constants.FieldVirtualMachineDeleteProtection)
}
