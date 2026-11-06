package importer

import (
	"reflect"
	"testing"
)

const (
	labelValueReady = "true"
	labelValueDev   = "dev"
)

// TestGetLabels checks that only the labels the user owns are imported into the
// `labels` attribute. Labels in the harvesterhci.io domain are maintained by
// Harvester and its controllers: importing them makes every subsequent plan
// propose their removal, and the controller writes them straight back.
func TestGetLabels(t *testing.T) {
	testcases := []struct {
		name     string
		labels   map[string]string
		expected map[string]string
	}{
		{
			name:     "no labels",
			labels:   map[string]string{},
			expected: map[string]string{},
		},
		{
			name: "controller labels of a VLAN network",
			labels: map[string]string{
				"network.harvesterhci.io/clusternetwork": "public",
				"network.harvesterhci.io/ready":          labelValueReady,
				"network.harvesterhci.io/type":           "L2VlanNetwork",
				"network.harvesterhci.io/vlan-id":        "2166",
			},
			expected: map[string]string{},
		},
		{
			name: "tags and harvester labels",
			labels: map[string]string{
				"tag.harvesterhci.io/environment": labelValueDev,
				"harvesterhci.io/creator":         "terraform",
			},
			expected: map[string]string{},
		},
		{
			name: "user labels are kept",
			labels: map[string]string{
				"environment":                            labelValueDev,
				"example.com/team":                       "network",
				"notharvesterhci.io/foo":                 "bar",
				"network.harvesterhci.io/clusternetwork": "public",
				"harvesterhci.io/creator":                "terraform",
			},
			expected: map[string]string{
				"environment":            labelValueDev,
				"example.com/team":       "network",
				"notharvesterhci.io/foo": "bar",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if labels := GetLabels(tc.labels); !reflect.DeepEqual(labels, tc.expected) {
				t.Errorf("imported labels = %v, want %v", labels, tc.expected)
			}
		})
	}
}
