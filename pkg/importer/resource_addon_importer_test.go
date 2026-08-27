package importer

import (
	"testing"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
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
		})
	}
}
