package kubeovn_subnet

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// Sentinel the SDK uses for values that are only known after apply.
const unknownConfigValue = "74D93920-ED26-11E3-AC10-0800200C9A66"

func subnetConfig(gateway string, excludeIPs []interface{}) map[string]interface{} {
	config := map[string]interface{}{
		constants.FieldCommonName:             "test-subnet",
		constants.FieldKubeOVNSubnetCIDRBlock: "10.0.0.0/24",
		constants.FieldKubeOVNSubnetGateway:   gateway,
	}
	if excludeIPs != nil {
		config[constants.FieldKubeOVNSubnetExcludeIPs] = excludeIPs
	}
	return config
}

func excludeIPsHash(ip string) string {
	set := ResourceKubeOVNSubnet().Schema[constants.FieldKubeOVNSubnetExcludeIPs].ZeroValue().(*schema.Set)
	return strconv.Itoa(set.F(ip))
}

// subnetState mimics a state refreshed from kube-ovn, where exclude_ips is
// already normalized.
func subnetState(excludeIPs ...string) *terraform.InstanceState {
	attributes := map[string]string{
		"id":                                          "test-subnet",
		constants.FieldCommonName:                     "test-subnet",
		constants.FieldKubeOVNSubnetCIDRBlock:         "10.0.0.0/24",
		constants.FieldKubeOVNSubnetGateway:           "10.0.0.1",
		constants.FieldKubeOVNSubnetExcludeIPs + ".#": strconv.Itoa(len(excludeIPs)),
	}
	for _, ip := range excludeIPs {
		attributes[constants.FieldKubeOVNSubnetExcludeIPs+"."+excludeIPsHash(ip)] = ip
	}
	return &terraform.InstanceState{ID: "test-subnet", Attributes: attributes}
}

// plannedExcludeIPs applies the diff the way the plugin server builds the
// planned state and returns the resulting exclude_ips, or computed=true when
// the attribute is only known after apply.
func plannedExcludeIPs(t *testing.T, state *terraform.InstanceState, config map[string]interface{}) (values []string, computed bool) {
	t.Helper()
	resource := ResourceKubeOVNSubnet()
	diff, err := resource.Diff(context.Background(), state, terraform.NewResourceConfigRaw(config), nil)
	if err != nil {
		t.Fatalf("unexpected diff error: %v", err)
	}
	if diff == nil {
		diff = &terraform.InstanceDiff{}
	}
	var prior map[string]string
	if state != nil {
		prior = state.Attributes
	}
	planned, err := diff.Apply(prior, resource.CoreConfigSchema())
	if err != nil {
		t.Fatalf("unexpected apply error: %v", err)
	}
	prefix := constants.FieldKubeOVNSubnetExcludeIPs + "."
	for key, value := range planned {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		if value == unknownConfigValue {
			return nil, true
		}
		if !strings.HasSuffix(key, ".#") {
			values = append(values, value)
		}
	}
	sort.Strings(values)
	return values, false
}

func TestPlanExcludeIPsReservesGateway(t *testing.T) {
	testcases := []struct {
		name             string
		gateway          string
		excludeIPs       []interface{}
		expected         []string
		expectedComputed bool
	}{
		{
			name:     "exclude_ips omitted",
			gateway:  "10.0.0.1",
			expected: []string{"10.0.0.1"},
		},
		{
			name:       "exclude_ips explicitly empty",
			gateway:    "10.0.0.1",
			excludeIPs: []interface{}{},
			expected:   []string{"10.0.0.1"},
		},
		{
			name:       "gateway missing from exclude_ips",
			gateway:    "10.0.0.1",
			excludeIPs: []interface{}{"10.0.0.50"},
			expected:   []string{"10.0.0.1", "10.0.0.50"},
		},
		{
			name:       "gateway already listed",
			gateway:    "10.0.0.1",
			excludeIPs: []interface{}{"10.0.0.50", "10.0.0.1"},
			expected:   []string{"10.0.0.1", "10.0.0.50"},
		},
		{
			name:       "range covering the gateway",
			gateway:    "10.0.0.1",
			excludeIPs: []interface{}{"10.0.0.1..10.0.0.10"},
			expected:   []string{"10.0.0.1..10.0.0.10"},
		},
		{
			name:             "gateway known after apply",
			gateway:          unknownConfigValue,
			excludeIPs:       []interface{}{"10.0.0.50"},
			expectedComputed: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			values, computed := plannedExcludeIPs(t, nil, subnetConfig(tc.gateway, tc.excludeIPs))
			if computed != tc.expectedComputed {
				t.Fatalf("expected computed=%v, got %v (values %v)", tc.expectedComputed, computed, values)
			}
			if !tc.expectedComputed && strings.Join(values, ",") != strings.Join(tc.expected, ",") {
				t.Errorf("expected planned exclude_ips %v, got %v", tc.expected, values)
			}
		})
	}
}

func TestPlanExcludeIPsWithUnknownEntry(t *testing.T) {
	// The plugin server hands the raw configuration to CustomizeDiff through
	// the prior state, which is how unknown set elements are detected.
	state := &terraform.InstanceState{
		RawConfig: cty.ObjectVal(map[string]cty.Value{
			constants.FieldKubeOVNSubnetExcludeIPs: cty.SetVal([]cty.Value{cty.UnknownVal(cty.String)}),
		}),
	}
	values, computed := plannedExcludeIPs(t, state, subnetConfig("10.0.0.1", []interface{}{unknownConfigValue}))
	if !computed {
		t.Errorf("expected exclude_ips to be known after apply, got %v", values)
	}
}

// Once kube-ovn has stored the normalized list, a configuration that does
// not repeat the gateway (or omits exclude_ips entirely) must plan no change.
func TestPlanExcludeIPsStableAfterNormalization(t *testing.T) {
	testcases := []struct {
		name       string
		stored     []string
		excludeIPs []interface{}
	}{
		{
			name:       "exclude_ips omitted",
			stored:     []string{"10.0.0.1"},
			excludeIPs: nil,
		},
		{
			name:       "gateway not repeated in the configuration",
			stored:     []string{"10.0.0.1", "10.0.0.50"},
			excludeIPs: []interface{}{"10.0.0.50"},
		},
		{
			name:       "configuration order differs from the stored order",
			stored:     []string{"10.0.0.1", "10.0.0.50", "10.0.0.60"},
			excludeIPs: []interface{}{"10.0.0.60", "10.0.0.50", "10.0.0.1"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			values, computed := plannedExcludeIPs(t, subnetState(tc.stored...), subnetConfig("10.0.0.1", tc.excludeIPs))
			if computed {
				t.Fatal("exclude_ips unexpectedly known after apply")
			}
			if strings.Join(values, ",") != strings.Join(tc.stored, ",") {
				t.Errorf("expected exclude_ips to stay %v, got %v", tc.stored, values)
			}
		})
	}
}

func TestPlanExcludeIPsAppliesConfigurationChanges(t *testing.T) {
	values, computed := plannedExcludeIPs(t, subnetState("10.0.0.1", "10.0.0.50"), subnetConfig("10.0.0.1", []interface{}{"10.0.0.60"}))
	if computed {
		t.Fatal("exclude_ips unexpectedly known after apply")
	}
	if strings.Join(values, ",") != "10.0.0.1,10.0.0.60" {
		t.Errorf("expected 10.0.0.50 replaced by 10.0.0.60 with the gateway kept, got %v", values)
	}
}
