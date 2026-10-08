package util

import (
	"maps"
	"testing"
)

func TestLabelsKeepsManagedLabels(t *testing.T) {
	const (
		tierKey   = "tier"
		tierValue = "fast"
	)

	testcases := []struct {
		name     string
		existing map[string]string
		config   map[string]interface{}
		expected map[string]string
	}{
		{
			name:     "labels of a new object",
			existing: nil,
			config:   map[string]interface{}{tierKey: tierValue, "team": "net"},
			expected: map[string]string{tierKey: tierValue, "team": "net"},
		},
		{
			name: "managed labels are kept and stale user labels dropped",
			existing: map[string]string{
				"tag.harvesterhci.io/role":          "db",
				"harvesterhci.io/creator":           "terraform-provider-harvester",
				"ovn.kubernetes.io/vpc-nat-gw-name": "gw",
				"ovn.kubernetes.io/eip_v4_ip":       "172.16.2.66",
				"stale-user-label":                  "x",
			},
			config: map[string]interface{}{tierKey: tierValue},
			expected: map[string]string{
				"tag.harvesterhci.io/role":          "db",
				"harvesterhci.io/creator":           "terraform-provider-harvester",
				"ovn.kubernetes.io/vpc-nat-gw-name": "gw",
				"ovn.kubernetes.io/eip_v4_ip":       "172.16.2.66",
				tierKey:                             tierValue,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			labels := maps.Clone(tc.existing)
			processors := NewProcessors().Labels(&labels)
			if err := processors[0].Parser(tc.config); err != nil {
				t.Fatal(err)
			}
			if !maps.Equal(labels, tc.expected) {
				t.Errorf("labels = %v, want %v", labels, tc.expected)
			}
		})
	}
}
