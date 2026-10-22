package kubeovn_subnet

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	kubeovnfake "github.com/kubeovn/kube-ovn/pkg/client/clientset/versioned/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	lowerRange      = "10.0.0.20..10.0.0.25"
	ipv6Range       = "fd00::1..fd00::ff"
	garbageRange    = "10.0.0.1..garbage"
)

func subnetConfig(gateway string, excludeIPs []interface{}) map[string]interface{} {
	config := map[string]interface{}{
		constants.FieldCommonName:             subnetName,
		constants.FieldKubeOVNSubnetCIDRBlock: subnetCIDR,
	}
	if gateway != "" {
		config[constants.FieldKubeOVNSubnetGateway] = gateway
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
			name:             "gateway omitted is computed by kube-ovn",
			gateway:          "",
			excludeIPs:       []interface{}{excludedIP},
			expectedComputed: true,
		},
		{
			name:       "dual-stack gateway reserves both addresses",
			gateway:    gatewayIP + ",fd00::1",
			excludeIPs: []interface{}{excludedIP},
			expected:   []string{gatewayIP, excludedIP, "fd00::1"},
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
		removed    bool
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
		{
			name:       "overlapping ranges are planned merged",
			stored:     []string{gatewayIP, excludedRange},
			excludeIPs: []interface{}{lowerRange, "10.0.0.24..10.0.0.30"},
			expected:   []string{gatewayIP, excludedRange},
		},
		{
			name:       "exclude_ips removed from the configuration keeps only the gateway",
			stored:     []string{gatewayIP, excludedIP},
			excludeIPs: nil,
			removed:    true,
			expected:   []string{gatewayIP},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			state := subnetState(tc.stored...)
			if tc.removed {
				// The plugin server hands the raw configuration to
				// CustomizeDiff, which is how a removed attribute is told
				// apart from the value the SDK keeps for a computed one.
				state.RawConfig = cty.ObjectVal(map[string]cty.Value{
					constants.FieldKubeOVNSubnetExcludeIPs: cty.NullVal(cty.Set(cty.String)),
				})
			}
			values, computed := plannedExcludeIPs(t, state, subnetConfig(gatewayIP, tc.excludeIPs))
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

func TestValidateDualStack(t *testing.T) {
	testcases := []struct {
		name        string
		field       string
		value       string
		expectedErr bool
	}{
		{name: "IPv4 CIDR", field: constants.FieldKubeOVNSubnetCIDRBlock, value: subnetCIDR},
		{name: "IPv6 CIDR", field: constants.FieldKubeOVNSubnetCIDRBlock, value: "fd00:10::/64"},
		{name: "dual-stack CIDRs", field: constants.FieldKubeOVNSubnetCIDRBlock, value: subnetCIDR + ",fd00:10::/64"},
		{name: "two IPv4 CIDRs", field: constants.FieldKubeOVNSubnetCIDRBlock, value: subnetCIDR + ",10.1.0.0/24", expectedErr: true},
		{name: "three CIDRs", field: constants.FieldKubeOVNSubnetCIDRBlock, value: subnetCIDR + ",fd00:10::/64,10.1.0.0/24", expectedErr: true},
		{name: "address instead of a CIDR", field: constants.FieldKubeOVNSubnetCIDRBlock, value: gatewayIP, expectedErr: true},
		{name: "IPv4 gateway", field: constants.FieldKubeOVNSubnetGateway, value: gatewayIP},
		{name: "dual-stack gateways", field: constants.FieldKubeOVNSubnetGateway, value: gatewayIP + ", fd00:10::1"},
		{name: "two IPv6 gateways", field: constants.FieldKubeOVNSubnetGateway, value: "fd00:10::1,fd00:10::2", expectedErr: true},
		{name: "CIDR instead of a gateway", field: constants.FieldKubeOVNSubnetGateway, value: subnetCIDR, expectedErr: true},
		{name: "empty gateway", field: constants.FieldKubeOVNSubnetGateway, value: "", expectedErr: true},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, errs := Schema()[tc.field].ValidateFunc(tc.value, tc.field)
			if (len(errs) > 0) != tc.expectedErr {
				t.Errorf("expected error=%v, got %v", tc.expectedErr, errs)
			}
		})
	}
}

func TestSubnetRefresh(t *testing.T) {
	condition := func(conditionType kubeovnv1.ConditionType, status corev1.ConditionStatus) kubeovnv1.Condition {
		return kubeovnv1.Condition{Type: conditionType, Status: status}
	}

	testcases := []struct {
		name          string
		conditions    []kubeovnv1.Condition
		missing       bool
		expectedState string
		expectedErr   error
	}{
		{
			name:          "subnet deleted",
			missing:       true,
			expectedState: constants.StateCommonRemoved,
		},
		{
			name:          "not reconciled yet",
			expectedState: constants.StateCommonActive,
		},
		{
			name:          "ready",
			conditions:    []kubeovnv1.Condition{condition(kubeovnv1.Ready, corev1.ConditionTrue)},
			expectedState: constants.StateCommonReady,
		},
		{
			name:          "spec rejected by the controller",
			conditions:    []kubeovnv1.Condition{condition(kubeovnv1.Validated, corev1.ConditionFalse)},
			expectedState: constants.StateCommonError,
			expectedErr:   ErrSubnetRejected,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			client := kubeovnfake.NewSimpleClientset()
			if !tc.missing {
				client = kubeovnfake.NewSimpleClientset(&kubeovnv1.Subnet{
					ObjectMeta: metav1.ObjectMeta{Name: subnetName},
					Status:     kubeovnv1.SubnetStatus{Conditions: tc.conditions},
				})
			}
			_, state, err := subnetRefresh(context.Background(), client.KubeovnV1().Subnets(), subnetName)()
			if state != tc.expectedState {
				t.Errorf("state = %q, want %q", state, tc.expectedState)
			}
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("error = %v, want %v", err, tc.expectedErr)
			}
		})
	}
}
