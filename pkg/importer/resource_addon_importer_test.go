package importer

import (
	"testing"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"github.com/rancher/wrangler/v3/pkg/condition"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceAddonStateGetter(t *testing.T) {
	testcases := []struct {
		name            string
		addon           *harvsterv1.Addon
		expectedID      string
		expectedState   string
		expectedEnabled bool
		expectedValues  string
		expectedDesc    string
		expectedTags    map[string]string
	}{
		{
			name: "enabled addon with values and successful status",
			addon: &harvsterv1.Addon{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pcidevices-controller",
					Namespace: "harvester-system",
					Labels: map[string]string{
						"tag.harvesterhci.io/env": "production",
					},
					Annotations: map[string]string{
						"field.cattle.io/description": "PCI devices controller",
					},
				},
				Spec: harvsterv1.AddonSpec{
					Enabled:       true,
					ValuesContent: "key: value",
					Repo:          "https://charts.example.com",
					Chart:         "pcidevices",
					Version:       "0.1.0",
				},
				Status: harvsterv1.AddonStatus{
					Status: harvsterv1.AddonDeployed,
				},
			},
			expectedID:      helper.BuildID("harvester-system", "pcidevices-controller"),
			expectedState:   "AddonDeploySuccessful",
			expectedEnabled: true,
			expectedValues:  "key: value",
			expectedDesc:    "PCI devices controller",
			expectedTags:    map[string]string{"env": "production"},
		},
		{
			name: "disabled addon with empty values",
			addon: &harvsterv1.Addon{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vm-import-controller",
					Namespace: "harvester-system",
					Labels:    map[string]string{},
				},
				Spec: harvsterv1.AddonSpec{
					Enabled: false,
					Repo:    "https://charts.example.com",
					Chart:   "vm-import",
					Version: "0.2.0",
				},
				Status: harvsterv1.AddonStatus{
					Status: harvsterv1.AddonDisabled,
				},
			},
			expectedID:      helper.BuildID("harvester-system", "vm-import-controller"),
			expectedState:   "AddonDisabled",
			expectedEnabled: false,
			expectedValues:  "",
			expectedDesc:    "",
			expectedTags:    map[string]string{},
		},
		{
			name: "addon with nil labels and annotations",
			addon: &harvsterv1.Addon{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "harvester-seeder",
					Namespace: "harvester-system",
				},
				Spec: harvsterv1.AddonSpec{
					Enabled: false,
				},
				Status: harvsterv1.AddonStatus{
					Status: harvsterv1.AddonInitState,
				},
			},
			expectedID:      helper.BuildID("harvester-system", "harvester-seeder"),
			expectedState:   "",
			expectedEnabled: false,
			expectedValues:  "",
			expectedDesc:    "",
			expectedTags:    map[string]string{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceAddonStateGetter(tc.addon)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.Name != tc.addon.Name {
				t.Errorf("Name: expected %q, got %q", tc.addon.Name, getter.Name)
			}
			if getter.ResourceType != constants.ResourceTypeAddon {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeAddon, getter.ResourceType)
			}

			state := getter.States[constants.FieldCommonState].(string)
			if state != tc.expectedState {
				t.Errorf("State: expected %q, got %q", tc.expectedState, state)
			}

			enabled := getter.States[constants.FieldAddonEnabled].(bool)
			if enabled != tc.expectedEnabled {
				t.Errorf("Enabled: expected %v, got %v", tc.expectedEnabled, enabled)
			}

			values := getter.States[constants.FieldAddonValuesContent].(string)
			if values != tc.expectedValues {
				t.Errorf("ValuesContent: expected %q, got %q", tc.expectedValues, values)
			}

			desc := getter.States[constants.FieldCommonDescription].(string)
			if desc != tc.expectedDesc {
				t.Errorf("Description: expected %q, got %q", tc.expectedDesc, desc)
			}

			tags := getter.States[constants.FieldCommonTags].(map[string]string)
			if len(tags) != len(tc.expectedTags) {
				t.Errorf("Tags: expected %d, got %d: %v", len(tc.expectedTags), len(tags), tags)
			}
			for key, val := range tc.expectedTags {
				if tags[key] != val {
					t.Errorf("Tag %q: expected %q, got %q", key, val, tags[key])
				}
			}

			// repo/chart/version must round-trip now that they are
			// user-configurable for custom addons.
			if repo := getter.States[constants.FieldAddonRepo].(string); repo != tc.addon.Spec.Repo {
				t.Errorf("Repo: expected %q, got %q", tc.addon.Spec.Repo, repo)
			}
			if chart := getter.States[constants.FieldAddonChart].(string); chart != tc.addon.Spec.Chart {
				t.Errorf("Chart: expected %q, got %q", tc.addon.Spec.Chart, chart)
			}
			if version := getter.States[constants.FieldAddonVersion].(string); version != tc.addon.Spec.Version {
				t.Errorf("Version: expected %q, got %q", tc.addon.Spec.Version, version)
			}
		})
	}
}

// TestAddonMessage verifies the message surfaced in state follows the
// failure > in-progress > completed priority.
func TestAddonMessage(t *testing.T) {
	cond := func(t condition.Cond, status corev1.ConditionStatus, message string) harvsterv1.Condition {
		return harvsterv1.Condition{Type: t, Status: status, Message: message}
	}
	testcases := []struct {
		name     string
		addon    *harvsterv1.Addon
		expected string
	}{
		{
			name:     "no conditions yields empty message",
			addon:    &harvsterv1.Addon{},
			expected: "",
		},
		{
			name: "failed operation wins over other conditions",
			addon: &harvsterv1.Addon{
				Status: harvsterv1.AddonStatus{Conditions: []harvsterv1.Condition{
					cond(harvsterv1.AddonOperationInProgress, corev1.ConditionTrue, "installing"),
					cond(harvsterv1.AddonOperationFailed, corev1.ConditionTrue, "helm install failed"),
				}},
			},
			expected: "helm install failed",
		},
		{
			name: "in-progress message when no failure",
			addon: &harvsterv1.Addon{
				Status: harvsterv1.AddonStatus{Conditions: []harvsterv1.Condition{
					cond(harvsterv1.AddonOperationInProgress, corev1.ConditionTrue, "installing"),
				}},
			},
			expected: "installing",
		},
		{
			name: "completed message as fallback",
			addon: &harvsterv1.Addon{
				Status: harvsterv1.AddonStatus{Conditions: []harvsterv1.Condition{
					cond(harvsterv1.AddonOperationFailed, corev1.ConditionFalse, ""),
					cond(harvsterv1.AddonOperationCompleted, corev1.ConditionTrue, "addon deployed"),
				}},
			},
			expected: "addon deployed",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if got := addonMessage(tc.addon); got != tc.expected {
				t.Errorf("addonMessage() = %q, want %q", got, tc.expected)
			}
		})
	}
}

// TestAddonUserLabels verifies the addon.harvesterhci.io/* labels never leak
// into the user labels state (they would drift on every plan otherwise).
func TestAddonUserLabels(t *testing.T) {
	getter, err := ResourceAddonStateGetter(&harvsterv1.Addon{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tf-test-dummy",
			Namespace: "harvester-system",
			Labels: map[string]string{
				"addon.harvesterhci.io/experimental": "true",
				"team":                               "infra",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	labels := getter.States[constants.FieldCommonLabels].(map[string]string)
	if _, leaked := labels["addon.harvesterhci.io/experimental"]; leaked {
		t.Errorf("experimental label leaked into user labels: %v", labels)
	}
	if labels["team"] != "infra" {
		t.Errorf("user label lost: %v", labels)
	}
}
