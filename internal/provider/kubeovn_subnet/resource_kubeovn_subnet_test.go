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
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

const (
	// Sentinel the SDK uses for values that are only known after apply.
	unknownConfigValue = "74D93920-ED26-11E3-AC10-0800200C9A66"

	subnetName      = "test-subnet"
	subnetCIDR      = "10.0.0.0/24"
	gatewayIP       = "10.0.0.1"
	gatewayRange    = "10.0.0.1..10.0.0.10"
	excludedIP      = "10.0.0.50"
	otherExcludedIP = "10.0.0.60"
	excludedRange   = "10.0.0.20..10.0.0.30"
	invalidEntry    = "not-an-ip"
)

func subnetConfig(gateway string, excludeIPs []interface{}) map[string]interface{} {
	config := map[string]interface{}{
		constants.FieldCommonName:             subnetName,
		constants.FieldKubeOVNSubnetCIDRBlock: subnetCIDR,
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
		"id":                                          subnetName,
		constants.FieldCommonName:                     subnetName,
		constants.FieldKubeOVNSubnetCIDRBlock:         subnetCIDR,
		constants.FieldKubeOVNSubnetGateway:           gatewayIP,
		constants.FieldKubeOVNSubnetExcludeIPs + ".#": strconv.Itoa(len(excludeIPs)),
	}
	for _, ip := range excludeIPs {
		attributes[constants.FieldKubeOVNSubnetExcludeIPs+"."+excludeIPsHash(ip)] = ip
	}
	return &terraform.InstanceState{ID: subnetName, Attributes: attributes}
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
		state            *terraform.InstanceState
		gateway          string
		excludeIPs       []interface{}
		expected         []string
		expectedComputed bool
	}{
		{
			name:     "exclude_ips omitted", //nolint:goconst
			gateway:  gatewayIP,
			expected: []string{gatewayIP},
		},
		{
			name:       "exclude_ips explicitly empty",
			gateway:    gatewayIP,
			excludeIPs: []interface{}{},
			expected:   []string{gatewayIP},
		},
		{
			name:       "gateway missing from exclude_ips",
			gateway:    gatewayIP,
			excludeIPs: []interface{}{excludedIP},
			expected:   []string{gatewayIP, excludedIP},
		},
		{
			name:       "gateway missing from several entries",
			gateway:    gatewayIP,
			excludeIPs: []interface{}{otherExcludedIP, excludedIP, excludedRange},
			expected:   []string{gatewayIP, excludedRange, excludedIP, otherExcludedIP},
		},
		{
			name:       "gateway already listed",
			gateway:    gatewayIP,
			excludeIPs: []interface{}{excludedIP, gatewayIP},
			expected:   []string{gatewayIP, excludedIP},
		},
		{
			name:       "range covering the gateway",
			gateway:    gatewayIP,
			excludeIPs: []interface{}{gatewayRange},
			expected:   []string{gatewayRange},
		},
		{
			name:       "empty entries are dropped",
			gateway:    gatewayIP,
			excludeIPs: []interface{}{"", excludedIP},
			expected:   []string{gatewayIP, excludedIP},
		},
		{
			name:       "unparsable entry does not cover the gateway",
			gateway:    gatewayIP,
			excludeIPs: []interface{}{invalidEntry},
			expected:   []string{gatewayIP, invalidEntry},
		},
		{
			name:             "gateway known after apply",
			gateway:          unknownConfigValue,
			excludeIPs:       []interface{}{excludedIP},
			expectedComputed: true,
		},
		{
			// The plugin server hands the raw configuration to CustomizeDiff
			// through the prior state, which is how unknown set elements are
			// detected.
			name: "exclude_ips entry known after apply",
			state: &terraform.InstanceState{
				RawConfig: cty.ObjectVal(map[string]cty.Value{
					constants.FieldKubeOVNSubnetExcludeIPs: cty.SetVal([]cty.Value{cty.UnknownVal(cty.String)}),
				}),
			},
			gateway:          gatewayIP,
			excludeIPs:       []interface{}{unknownConfigValue},
			expectedComputed: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			values, computed := plannedExcludeIPs(t, tc.state, subnetConfig(tc.gateway, tc.excludeIPs))
			if computed != tc.expectedComputed {
				t.Fatalf("expected computed=%v, got %v (values %v)", tc.expectedComputed, computed, values)
			}
			if !tc.expectedComputed && strings.Join(values, ",") != strings.Join(tc.expected, ",") {
				t.Errorf("expected planned exclude_ips %v, got %v", tc.expected, values)
			}
		})
	}
}

// Once kube-ovn has stored the normalized list, a configuration that does
// not repeat the gateway (or omits exclude_ips entirely) must plan no change,
// while a real configuration change is still planned with the gateway kept.
func TestPlanExcludeIPsAgainstStoredState(t *testing.T) {
	testcases := []struct {
		name       string
		stored     []string
		excludeIPs []interface{}
		expected   []string
	}{
		{
			name:       "exclude_ips omitted", //nolint:goconst
			stored:     []string{gatewayIP},
			excludeIPs: nil,
			expected:   []string{gatewayIP},
		},
		{
			name:       "gateway not repeated in the configuration",
			stored:     []string{gatewayIP, excludedIP},
			excludeIPs: []interface{}{excludedIP},
			expected:   []string{gatewayIP, excludedIP},
		},
		{
			name:       "configuration order differs from the stored order",
			stored:     []string{gatewayIP, excludedIP, otherExcludedIP},
			excludeIPs: []interface{}{otherExcludedIP, excludedIP, gatewayIP},
			expected:   []string{gatewayIP, excludedIP, otherExcludedIP},
		},
		{
			name:       "changed entry is replaced and the gateway kept",
			stored:     []string{gatewayIP, excludedIP},
			excludeIPs: []interface{}{otherExcludedIP},
			expected:   []string{gatewayIP, otherExcludedIP},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			values, computed := plannedExcludeIPs(t, subnetState(tc.stored...), subnetConfig(gatewayIP, tc.excludeIPs))
			if computed {
				t.Fatal("exclude_ips unexpectedly known after apply")
			}
			if strings.Join(values, ",") != strings.Join(tc.expected, ",") {
				t.Errorf("expected exclude_ips %v, got %v", tc.expected, values)
			}
		})
	}
}

// kube-ovn derives the protocol from cidr_block and overwrites the spec, so an
// omitted protocol must neither be sent nor planned against the stored value.
func TestSubnetProtocol(t *testing.T) {
	const ipv6CIDR = "fd00:10::/64"

	testcases := []struct {
		name         string
		cidr         string
		stored       string // protocol in the state, empty for a new subnet
		configured   string // protocol in the configuration, empty when omitted
		expectedSent string
		expectedDiff bool
	}{
		{
			name:         "omitted on create is left to kube-ovn",
			cidr:         ipv6CIDR,
			expectedSent: "",
			expectedDiff: true, // computed, known after apply
		},
		{
			name:         "configured on create is sent",
			cidr:         ipv6CIDR,
			configured:   kubeovnv1.ProtocolIPv6,
			expectedSent: kubeovnv1.ProtocolIPv6,
			expectedDiff: true,
		},
		{
			name:         "omitted with an IPv6 subnet stored plans no change",
			cidr:         ipv6CIDR,
			stored:       kubeovnv1.ProtocolIPv6,
			expectedSent: kubeovnv1.ProtocolIPv6,
		},
		{
			name:         "omitted with an IPv4 subnet stored plans no change",
			cidr:         subnetCIDR,
			stored:       kubeovnv1.ProtocolIPv4,
			expectedSent: kubeovnv1.ProtocolIPv4,
		},
		{
			name:         "configured value different from the stored one is planned",
			cidr:         subnetCIDR,
			stored:       kubeovnv1.ProtocolIPv4,
			configured:   kubeovnv1.ProtocolIPv6,
			expectedSent: kubeovnv1.ProtocolIPv6,
			expectedDiff: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			config := map[string]interface{}{
				constants.FieldCommonName:             subnetName,
				constants.FieldKubeOVNSubnetCIDRBlock: tc.cidr,
				constants.FieldKubeOVNSubnetGateway:   gatewayIP,
			}
			if tc.configured != "" {
				config[constants.FieldKubeOVNSubnetProtocol] = tc.configured
			}
			var state *terraform.InstanceState
			if tc.stored != "" {
				state = subnetState(gatewayIP)
				state.Attributes[constants.FieldKubeOVNSubnetCIDRBlock] = tc.cidr
				state.Attributes[constants.FieldKubeOVNSubnetProtocol] = tc.stored
			}

			resource := ResourceKubeOVNSubnet()
			diff, err := resource.Diff(context.Background(), state, terraform.NewResourceConfigRaw(config), nil)
			if err != nil {
				t.Fatalf("unexpected diff error: %v", err)
			}
			_, planned := diff.GetAttribute(constants.FieldKubeOVNSubnetProtocol)
			if planned != tc.expectedDiff {
				t.Errorf("protocol planned = %v, want %v", planned, tc.expectedDiff)
			}

			var prior map[string]string
			if state != nil {
				prior = state.Attributes
			}
			d, err := schema.InternalMap(resource.Schema).Data(&terraform.InstanceState{Attributes: prior}, diff)
			if err != nil {
				t.Fatal(err)
			}
			constructor := Creator(subnetName)
			if state != nil {
				constructor = Updater(&kubeovnv1.Subnet{Spec: kubeovnv1.SubnetSpec{CIDRBlock: tc.cidr, Protocol: tc.stored}})
			}
			obj, err := util.ResourceConstruct(context.Background(), d, constructor)
			if err != nil {
				t.Fatal(err)
			}
			if got := obj.(*kubeovnv1.Subnet).Spec.Protocol; got != tc.expectedSent {
				t.Errorf("protocol sent = %q, want %q", got, tc.expectedSent)
			}
		})
	}
}

func TestSubnetReady(t *testing.T) {
	const message = "subnet cidr conflicts with subnet other"
	condition := func(conditionType kubeovnv1.ConditionType, status corev1.ConditionStatus) kubeovnv1.Condition {
		return kubeovnv1.Condition{Type: conditionType, Status: status, Message: message}
	}

	testcases := []struct {
		name          string
		conditions    []kubeovnv1.Condition
		expectedReady bool
		expectedErr   bool
	}{
		{
			name: "not reconciled yet",
		},
		{
			name:       "validated but not ready",
			conditions: []kubeovnv1.Condition{condition(kubeovnv1.Validated, corev1.ConditionTrue), condition(kubeovnv1.Ready, corev1.ConditionFalse)},
		},
		{
			name: "ready",
			conditions: []kubeovnv1.Condition{
				condition(kubeovnv1.Validated, corev1.ConditionTrue),
				condition(kubeovnv1.Ready, corev1.ConditionTrue),
				condition(kubeovnv1.Error, corev1.ConditionUnknown),
			},
			expectedReady: true,
		},
		{
			name:        "spec rejected by the controller",
			conditions:  []kubeovnv1.Condition{condition(kubeovnv1.Validated, corev1.ConditionFalse), condition(kubeovnv1.Ready, corev1.ConditionFalse)},
			expectedErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ready, err := subnetReady(&kubeovnv1.Subnet{Status: kubeovnv1.SubnetStatus{Conditions: tc.conditions}})
			if ready != tc.expectedReady {
				t.Errorf("ready = %v, want %v", ready, tc.expectedReady)
			}
			if tc.expectedErr != (err != nil) {
				t.Errorf("err = %v, want an error: %v", err, tc.expectedErr)
			}
			if err != nil && !strings.Contains(err.Error(), message) {
				t.Errorf("err = %v, want the controller message in it", err)
			}
		})
	}
}
