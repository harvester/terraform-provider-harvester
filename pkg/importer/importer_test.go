package importer

import (
	"maps"
	"testing"
)

func TestGetLabelsSkipsManagedLabels(t *testing.T) {
	testcases := []struct {
		name     string
		labels   map[string]string
		expected map[string]string
	}{
		{
			name:     "nil labels",
			labels:   nil,
			expected: map[string]string{},
		},
		{
			name: "only managed labels",
			labels: map[string]string{
				"tag.harvesterhci.io/role":          "db",
				"harvesterhci.io/creator":           "terraform-provider-harvester",
				"ovn.kubernetes.io/vpc-nat-gw-name": "gw",
			},
			expected: map[string]string{},
		},
		{
			name: "several user labels are kept next to managed ones",
			labels: map[string]string{
				"tag.harvesterhci.io/role":          "db",
				"harvesterhci.io/creator":           "terraform-provider-harvester",
				"ovn.kubernetes.io/vpc-nat-gw-name": "gw",
				"ovn.kubernetes.io/eip_v4_ip":       "172.16.2.66",
				"tier":                              "fast",
				"team":                              "net",
			},
			expected: map[string]string{"tier": "fast", "team": "net"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if got := GetLabels(tc.labels); !maps.Equal(got, tc.expected) {
				t.Errorf("GetLabels() = %v, want %v", got, tc.expected)
			}
		})
	}
}
