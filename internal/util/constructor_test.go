package util

import (
	"reflect"
	"testing"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// TestProcessorsLabels checks that constructing an update keeps the labels
// Harvester owns on the object while the ones the user manages are rebuilt
// from the configuration.
func TestProcessorsLabels(t *testing.T) {
	testcases := []struct {
		name       string
		labels     map[string]string
		configured map[string]interface{}
		expected   map[string]string
	}{
		{
			name: "controller labels survive an empty configuration",
			labels: map[string]string{
				"network.harvesterhci.io/clusternetwork": "public",
				"network.harvesterhci.io/ready":          "true",
				"network.harvesterhci.io/type":           "L2VlanNetwork",
				"network.harvesterhci.io/vlan-id":        "2166",
			},
			configured: map[string]interface{}{},
			expected: map[string]string{
				"network.harvesterhci.io/clusternetwork": "public",
				"network.harvesterhci.io/ready":          "true",
				"network.harvesterhci.io/type":           "L2VlanNetwork",
				"network.harvesterhci.io/vlan-id":        "2166",
			},
		},
		{
			name: "a removed user label is dropped",
			labels: map[string]string{
				"network.harvesterhci.io/vlan-id": "2166",
				"environment":                     "dev",
			},
			configured: map[string]interface{}{},
			expected: map[string]string{
				"network.harvesterhci.io/vlan-id": "2166",
			},
		},
		{
			name: "a configured user label is applied",
			labels: map[string]string{
				"network.harvesterhci.io/vlan-id": "2166",
				"environment":                     "dev",
			},
			configured: map[string]interface{}{
				"environment": "prod",
			},
			expected: map[string]string{
				"network.harvesterhci.io/vlan-id": "2166",
				"environment":                     "prod",
			},
		},
		{
			name:   "a configured reserved label is ignored",
			labels: map[string]string{},
			configured: map[string]interface{}{
				"network.harvesterhci.io/ready": "false",
				"harvesterhci.io/creator":       "me",
				"environment":                   "dev",
			},
			expected: map[string]string{
				"environment": "dev",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			labels := tc.labels
			processors := NewProcessors().Labels(&labels)

			var parser func(interface{}) error
			for _, processor := range processors {
				if processor.Field == constants.FieldCommonLabels {
					parser = processor.Parser
				}
			}
			if parser == nil {
				t.Fatalf("no processor registered for %s", constants.FieldCommonLabels)
			}
			if err := parser(tc.configured); err != nil {
				t.Fatalf("parsing the configured labels: %v", err)
			}

			if !reflect.DeepEqual(labels, tc.expected) {
				t.Errorf("constructed labels = %v, want %v", labels, tc.expected)
			}
		})
	}
}
