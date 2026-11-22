package kubeovn_vlan

import (
	"context"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// TestVlanDiff checks which changes replace the VLAN: kube-ovn retags the
// subnets in place when the ID changes, but never moves them to another
// provider network.
func TestVlanDiff(t *testing.T) {
	state := &terraform.InstanceState{
		ID: vlanName,
		Attributes: map[string]string{
			"id":                               vlanName,
			constants.FieldCommonName:          vlanName,
			constants.FieldKubeOVNVlanID:       strconv.Itoa(vlanID),
			constants.FieldKubeOVNVlanProvider: vlanProvider,
			// Computed attributes, as refreshed from kube-ovn.
			constants.FieldKubeOVNVlanStatusSubnets + ".#": "0",
			constants.FieldKubeOVNVlanStatusConflict:       "false",
			constants.FieldCommonState:                     constants.StateCommonActive,
		},
	}

	testcases := []struct {
		name            string
		id              int
		provider        string
		expectedChange  string // attribute expected in the diff, "" for none
		expectedReplace bool
	}{
		{name: "no change", id: vlanID, provider: vlanProvider},
		{name: "vlan_id changed in place", id: otherVlanID, provider: vlanProvider, expectedChange: constants.FieldKubeOVNVlanID},
		{name: "vlan_id set to untagged in place", id: 0, provider: vlanProvider, expectedChange: constants.FieldKubeOVNVlanID},
		{name: "network_provider changed", id: vlanID, provider: "other-pn", expectedChange: constants.FieldKubeOVNVlanProvider, expectedReplace: true},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			config := terraform.NewResourceConfigRaw(map[string]interface{}{
				constants.FieldCommonName:          vlanName,
				constants.FieldKubeOVNVlanID:       tc.id,
				constants.FieldKubeOVNVlanProvider: tc.provider,
			})
			diff, err := ResourceKubeOVNVlan().Diff(context.Background(), state, config, nil)
			if err != nil {
				t.Fatalf("unexpected diff error: %v", err)
			}
			if tc.expectedChange == "" {
				if diff != nil && !diff.Empty() {
					t.Fatalf("expected no diff, got %v", diff.Attributes)
				}
				return
			}
			if diff == nil || diff.Attributes[tc.expectedChange] == nil {
				t.Fatalf("expected a change of %s, got %v", tc.expectedChange, diff)
			}
			if diff.RequiresNew() != tc.expectedReplace {
				t.Errorf("RequiresNew() = %v, want %v", diff.RequiresNew(), tc.expectedReplace)
			}
		})
	}
}
