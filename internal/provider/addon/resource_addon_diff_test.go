package addon

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func addonDiff(t *testing.T, enabled string) *terraform.InstanceDiff {
	t.Helper()
	state := &terraform.InstanceState{
		ID: "harvester-system/vm-import-controller",
		Attributes: map[string]string{
			"id":                           "harvester-system/vm-import-controller",
			constants.FieldCommonName:      "vm-import-controller",
			constants.FieldCommonNamespace: "harvester-system",
			constants.FieldAddonEnabled:    "true",
			constants.FieldCommonState:     "AddonDeploySuccessful",
			constants.FieldCommonMessage:   "",
		},
	}
	config := map[string]interface{}{
		constants.FieldCommonName:      "vm-import-controller",
		constants.FieldCommonNamespace: "harvester-system",
		constants.FieldAddonEnabled:    enabled,
	}
	diff, err := ResourceAddon().Diff(context.Background(), state, terraform.NewResourceConfigRaw(config), nil)
	if err != nil {
		t.Fatalf("unexpected diff error: %v", err)
	}
	return diff
}

// Toggling the addon starts an operation whose outcome is only known after
// apply: state and message must be planned as unknown so that outputs built
// from them do not keep the value from before the operation.
func TestAddonToggleMarksStatusUnknown(t *testing.T) {
	diff := addonDiff(t, "false")
	if diff == nil {
		t.Fatal("expected a diff when enabled changes")
	}
	for _, key := range []string{constants.FieldCommonState, constants.FieldCommonMessage} {
		attr := diff.Attributes[key]
		if attr == nil || !attr.NewComputed {
			t.Errorf("expected %s to be known after apply, got %+v", key, attr)
		}
	}
}

func TestAddonUnchangedKeepsStatus(t *testing.T) {
	diff := addonDiff(t, "true")
	if diff == nil {
		return
	}
	for _, key := range []string{constants.FieldCommonState, constants.FieldCommonMessage} {
		if attr := diff.Attributes[key]; attr != nil && attr.NewComputed {
			t.Errorf("%s must keep its value when nothing triggers an operation", key)
		}
	}
}
