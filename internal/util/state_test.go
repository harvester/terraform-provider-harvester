package util

import (
	"reflect"
	"testing"
)

// TestManagedLabels checks which labels of an object reach the state of a
// resource: the ones Terraform already knows about, so that an update never
// removes a label written by Harvester, by another controller or by hand.
func TestManagedLabels(t *testing.T) {
	testcases := []struct {
		name     string
		known    interface{}
		actual   interface{}
		expected interface{}
	}{
		{
			name:  "labels nobody declared are not recorded",
			known: map[string]interface{}{},
			actual: map[string]string{
				labelKeyClusterNetwork: "public",
				labelKeyKubevirt:       "vm",
				"owner":                "someone-else",
			},
			expected: map[string]string{},
		},
		{
			name: "a declared label is recorded with the value of the object",
			known: map[string]interface{}{
				labelKeyEnvironment: labelValueDev,
			},
			actual: map[string]string{
				labelKeyEnvironment:    "prod",
				labelKeyClusterNetwork: "public",
			},
			expected: map[string]string{
				labelKeyEnvironment: "prod",
			},
		},
		{
			name: "a declared label missing from the object is not recorded",
			known: map[string]interface{}{
				labelKeyEnvironment: labelValueDev,
			},
			actual: map[string]string{
				labelKeyClusterNetwork: "public",
			},
			expected: map[string]string{},
		},
		{
			name:     "a value of an unexpected type is left alone",
			known:    "not a map",
			actual:   map[string]string{labelKeyEnvironment: labelValueDev},
			expected: map[string]string{labelKeyEnvironment: labelValueDev},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if managed := managedLabels(tc.known, tc.actual); !reflect.DeepEqual(managed, tc.expected) {
				t.Errorf("managedLabels() = %v, want %v", managed, tc.expected)
			}
		})
	}
}
