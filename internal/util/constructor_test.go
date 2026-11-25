package util

import (
	"reflect"
	"testing"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

const (
	labelKeyClusterNetwork = "network.harvesterhci.io/clusternetwork"
	labelKeyReady          = "network.harvesterhci.io/ready"
	labelKeyType           = "network.harvesterhci.io/type"
	labelKeyVlanID         = "network.harvesterhci.io/vlan-id"
	labelKeyKubevirt       = "kubevirt.io/domain"
	labelKeyEnvironment    = "environment"
	labelValueDev          = "dev"
)

// TestProcessorsLabels checks that constructing an update applies the labels
// of the configuration and leaves every label Terraform does not manage on the
// object, whoever wrote it. A label is removed only when the state carried it
// and the configuration dropped it.
func TestProcessorsLabels(t *testing.T) {
	testcases := []struct {
		name       string
		labels     map[string]string
		state      map[string]interface{}
		configured map[string]interface{}
		expected   map[string]string
	}{
		{
			name: "controller labels survive an empty configuration",
			labels: map[string]string{
				labelKeyClusterNetwork: "public",
				labelKeyReady:          "true",
				labelKeyType:           "L2VlanNetwork",
				labelKeyVlanID:         "2166",
			},
			state:      map[string]interface{}{},
			configured: map[string]interface{}{},
			expected: map[string]string{
				labelKeyClusterNetwork: "public",
				labelKeyReady:          "true",
				labelKeyType:           "L2VlanNetwork",
				labelKeyVlanID:         "2166",
			},
		},
		{
			name: "labels of another domain are left alone",
			labels: map[string]string{
				labelKeyVlanID:   "2166",
				labelKeyKubevirt: "vm",
			},
			state:      map[string]interface{}{},
			configured: map[string]interface{}{},
			expected: map[string]string{
				labelKeyVlanID:   "2166",
				labelKeyKubevirt: "vm",
			},
		},
		{
			name: "a label added out of band is left alone",
			labels: map[string]string{
				labelKeyVlanID:      "2166",
				labelKeyEnvironment: labelValueDev,
				"owner":             "someone-else",
			},
			state: map[string]interface{}{
				labelKeyEnvironment: labelValueDev,
			},
			configured: map[string]interface{}{
				labelKeyEnvironment: labelValueDev,
			},
			expected: map[string]string{
				labelKeyVlanID:      "2166",
				labelKeyEnvironment: labelValueDev,
				"owner":             "someone-else",
			},
		},
		{
			name: "a user label removed from the configuration is dropped",
			labels: map[string]string{
				labelKeyVlanID:      "2166",
				labelKeyEnvironment: labelValueDev,
			},
			state: map[string]interface{}{
				labelKeyEnvironment: labelValueDev,
			},
			configured: map[string]interface{}{},
			expected: map[string]string{
				labelKeyVlanID: "2166",
			},
		},
		{
			name: "a configured user label is applied",
			labels: map[string]string{
				labelKeyVlanID:      "2166",
				labelKeyEnvironment: labelValueDev,
			},
			state: map[string]interface{}{
				labelKeyEnvironment: labelValueDev,
			},
			configured: map[string]interface{}{
				labelKeyEnvironment: "prod",
			},
			expected: map[string]string{
				labelKeyVlanID:      "2166",
				labelKeyEnvironment: "prod",
			},
		},
		{
			name:  "a configured reserved label is ignored",
			state: map[string]interface{}{},
			labels: map[string]string{
				labelKeyReady: "true",
			},
			configured: map[string]interface{}{
				labelKeyReady:             "false",
				"harvesterhci.io/creator": "me",
				labelKeyEnvironment:       labelValueDev,
			},
			expected: map[string]string{
				labelKeyReady:       "true",
				labelKeyEnvironment: labelValueDev,
			},
		},
		{
			name: "a reserved label left in the state is not removed",
			labels: map[string]string{
				labelKeyReady: "true",
			},
			state: map[string]interface{}{
				labelKeyReady: "true",
			},
			configured: map[string]interface{}{},
			expected: map[string]string{
				labelKeyReady: "true",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			labels := tc.labels
			processors := NewProcessors().Labels(&labels)

			var parser func(interface{}, interface{}) error
			for _, processor := range processors {
				if processor.Field == constants.FieldCommonLabels {
					parser = processor.ChangeParser
				}
			}
			if parser == nil {
				t.Fatalf("no processor registered for %s", constants.FieldCommonLabels)
			}
			if err := parser(tc.state, tc.configured); err != nil {
				t.Fatalf("parsing the configured labels: %v", err)
			}

			if !reflect.DeepEqual(labels, tc.expected) {
				t.Errorf("constructed labels = %v, want %v", labels, tc.expected)
			}
		})
	}
}

// TestProcessorsLabelsDoesNotMutateOnSetup checks that registering the
// processor leaves the object alone: the labels are only touched once the
// configuration and the state are known, so a resource whose update fails
// before the processors run keeps its labels.
func TestProcessorsLabelsDoesNotMutateOnSetup(t *testing.T) {
	labels := map[string]string{
		labelKeyVlanID:      "2166",
		labelKeyEnvironment: labelValueDev,
	}
	expected := map[string]string{
		labelKeyVlanID:      "2166",
		labelKeyEnvironment: labelValueDev,
	}

	NewProcessors().Labels(&labels)

	if !reflect.DeepEqual(labels, expected) {
		t.Errorf("labels after registering the processor = %v, want %v", labels, expected)
	}
}
