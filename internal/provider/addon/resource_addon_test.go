package addon

import (
	"strings"
	"testing"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func addonWithState(status harvsterv1.AddonState, conditions ...harvsterv1.Condition) *harvsterv1.Addon {
	return &harvsterv1.Addon{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-addon",
			Namespace: "harvester-system",
		},
		Status: harvsterv1.AddonStatus{
			Status:     status,
			Conditions: conditions,
		},
	}
}

// TestAddonCurrentState verifies the state mapping used by the wait loops:
// every CRD state maps to a non-empty StateChangeConf state, and a failed
// operation condition surfaces its message as an error.
func TestAddonCurrentState(t *testing.T) {
	testcases := []struct {
		name          string
		addon         *harvsterv1.Addon
		expectedState string
		expectError   string
	}{
		{
			name:          "init state maps to the non-empty sentinel",
			addon:         addonWithState(harvsterv1.AddonInitState),
			expectedState: constants.StateAddonInit,
		},
		{
			name:          "enabling",
			addon:         addonWithState(harvsterv1.AddonEnabling),
			expectedState: string(harvsterv1.AddonEnabling),
		},
		{
			name:          "deployed",
			addon:         addonWithState(harvsterv1.AddonDeployed),
			expectedState: string(harvsterv1.AddonDeployed),
		},
		{
			name:          "updating",
			addon:         addonWithState(harvsterv1.AddonUpdating),
			expectedState: string(harvsterv1.AddonUpdating),
		},
		{
			name:          "disabling",
			addon:         addonWithState(harvsterv1.AddonDisabling),
			expectedState: string(harvsterv1.AddonDisabling),
		},
		{
			name:          "disabled",
			addon:         addonWithState(harvsterv1.AddonDisabled),
			expectedState: string(harvsterv1.AddonDisabled),
		},
		{
			name: "failed operation surfaces the controller message",
			addon: addonWithState(harvsterv1.AddonEnabling, harvsterv1.Condition{
				Type:    harvsterv1.AddonOperationFailed,
				Status:  corev1.ConditionTrue,
				Message: "helm install failed: chart not found",
			}),
			expectError: "helm install failed: chart not found",
		},
		{
			name: "explicit false failed condition is not an error",
			addon: addonWithState(harvsterv1.AddonDeployed, harvsterv1.Condition{
				Type:   harvsterv1.AddonOperationFailed,
				Status: corev1.ConditionFalse,
			}),
			expectedState: string(harvsterv1.AddonDeployed),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			state, err := addonCurrentState(tc.addon)
			if tc.expectError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectError) {
					t.Fatalf("addonCurrentState() error = %v, want containing %q", err, tc.expectError)
				}
				return
			}
			if err != nil {
				t.Fatalf("addonCurrentState() unexpected error: %v", err)
			}
			if state != tc.expectedState {
				t.Errorf("addonCurrentState() = %q, want %q", state, tc.expectedState)
			}
		})
	}
}
