package util

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func labelTestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		constants.FieldCommonName: {
			Type:     schema.TypeString,
			Required: true,
		},
		constants.FieldCommonLabels: {
			Type:     schema.TypeMap,
			Optional: true,
		},
	}
}

// TestStatesSetLabels checks which labels of an object are written to the
// state, for a resource that Terraform already knows, for the read that
// follows a terraform import, and for a data source.
func TestStatesSetLabels(t *testing.T) {
	const (
		name       = "public-vlan-2166"
		foreignKey = "example.com/owner"
	)
	onTheObject := map[string]string{
		labelKeyEnvironment: "prod",
		foreignKey:          "someone-else",
	}

	testcases := []struct {
		name       string
		prior      map[string]interface{}
		dataSetter func(*schema.ResourceData, *importer.StateGetter) error
		expected   map[string]interface{}
	}{
		{
			name: "a known resource records only the labels it manages",
			prior: map[string]interface{}{
				constants.FieldCommonName:   name,
				constants.FieldCommonLabels: map[string]interface{}{labelKeyEnvironment: "dev"},
			},
			dataSetter: ResourceStatesSet,
			expected:   map[string]interface{}{labelKeyEnvironment: "prod"},
		},
		{
			name: "a known resource with no labels records none",
			prior: map[string]interface{}{
				constants.FieldCommonName: name,
			},
			dataSetter: ResourceStatesSet,
			expected:   map[string]interface{}{},
		},
		{
			name:       "an imported resource adopts the labels of the object",
			prior:      map[string]interface{}{},
			dataSetter: ResourceStatesSet,
			expected: map[string]interface{}{
				labelKeyEnvironment: "prod",
				foreignKey:          "someone-else",
			},
		},
		{
			name: "a data source reports every label",
			prior: map[string]interface{}{
				constants.FieldCommonName: name,
			},
			dataSetter: DataSourceStatesSet,
			expected: map[string]interface{}{
				labelKeyEnvironment: "prod",
				foreignKey:          "someone-else",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, labelTestSchema(), tc.prior)
			getter := &importer.StateGetter{
				ID: "harvester-public/" + name,
				States: map[string]interface{}{
					// The name is set in the same pass as the labels, so this
					// also covers the case where it is written first.
					constants.FieldCommonName:   name,
					constants.FieldCommonLabels: onTheObject,
				},
			}

			if err := tc.dataSetter(d, getter); err != nil {
				t.Fatalf("setting the states: %v", err)
			}

			if labels := d.Get(constants.FieldCommonLabels); !reflect.DeepEqual(labels, tc.expected) {
				t.Errorf("labels in the state = %v, want %v", labels, tc.expected)
			}
		})
	}
}

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
