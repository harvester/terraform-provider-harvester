package kubeovn_qos_policy

import (
	"context"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

const (
	qosName = "test-qos"
)

func bandwidthRule(name, rateMax string) map[string]interface{} {
	return map[string]interface{}{
		constants.FieldKubeOVNQoSRuleName:      name,
		constants.FieldKubeOVNQoSRuleRateMax:   rateMax,
		constants.FieldKubeOVNQoSRuleDirection: string(kubeovnv1.QoSDirectionIngress),
	}
}

func TestQoSPolicyConstructor(t *testing.T) {
	const (
		rule1 = "rule1"
		rule2 = "rule2"
	)

	existing := &kubeovnv1.QoSPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: qosName},
		Spec: kubeovnv1.QoSPolicySpec{
			BindingType:         kubeovnv1.QoSBindingTypeEIP,
			BandwidthLimitRules: kubeovnv1.QoSPolicyBandwidthLimitRules{{Name: "old-rule"}},
		},
	}

	testcases := []struct {
		name          string
		existing      *kubeovnv1.QoSPolicy // nil creates the policy
		config        map[string]interface{}
		expectedRules []string // names, in order
	}{
		{
			name: "create keeps the order of several rules",
			config: map[string]interface{}{
				constants.FieldKubeOVNQoSBindingType:         string(kubeovnv1.QoSBindingTypeNatGw),
				constants.FieldKubeOVNQoSBandwidthLimitRules: []interface{}{bandwidthRule(rule2, "20M"), bandwidthRule(rule1, "10M"), bandwidthRule("rule3", "30M")},
			},
			expectedRules: []string{rule2, rule1, "rule3"},
		},
		{
			name:     "update drops rules removed from the configuration",
			existing: existing,
			config:   map[string]interface{}{},
		},
		{
			name:     "update replaces rules instead of appending to them",
			existing: existing,
			config: map[string]interface{}{
				constants.FieldKubeOVNQoSBandwidthLimitRules: []interface{}{bandwidthRule(rule1, "10M"), bandwidthRule(rule2, "20M")},
			},
			expectedRules: []string{rule1, rule2},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			constructor := Creator(qosName)
			if tc.existing != nil {
				constructor = Updater(tc.existing.DeepCopy())
			}
			d := schema.TestResourceDataRaw(t, Schema(), tc.config)
			result, err := util.ResourceConstruct(context.Background(), d, constructor)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			qos := result.(*kubeovnv1.QoSPolicy)

			rules := make([]string, 0, len(qos.Spec.BandwidthLimitRules))
			for _, rule := range qos.Spec.BandwidthLimitRules {
				rules = append(rules, rule.Name)
			}
			if !slices.Equal(rules, tc.expectedRules) {
				t.Errorf("BandwidthLimitRules: expected %v, got %v", tc.expectedRules, rules)
			}
		})
	}
}

// TestQoSPolicyConstructorShared checks that shared is sent as configured, false included, so
// that an update can switch it off.
func TestQoSPolicyConstructorShared(t *testing.T) {
	testcases := []struct {
		name     string
		existing *kubeovnv1.QoSPolicy // nil creates the object
		value    bool
	}{
		{name: "create with true", value: true},
		{name: "create with false", value: false},
		{name: "update switches it off", existing: &kubeovnv1.QoSPolicy{ObjectMeta: metav1.ObjectMeta{Name: qosName}, Spec: kubeovnv1.QoSPolicySpec{Shared: true, BindingType: kubeovnv1.QoSBindingTypeEIP}}, value: false},
		{name: "update switches it on", existing: &kubeovnv1.QoSPolicy{ObjectMeta: metav1.ObjectMeta{Name: qosName}, Spec: kubeovnv1.QoSPolicySpec{BindingType: kubeovnv1.QoSBindingTypeEIP}}, value: true},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, Schema(), map[string]interface{}{
				constants.FieldCommonName:            qosName,
				constants.FieldKubeOVNQoSBindingType: string(kubeovnv1.QoSBindingTypeEIP),
				constants.FieldKubeOVNQoSShared:      tc.value,
			})
			constructor := Creator(qosName)
			if tc.existing != nil {
				constructor = Updater(tc.existing)
			}
			obj, err := util.ResourceConstruct(context.Background(), d, constructor)
			if err != nil {
				t.Fatal(err)
			}
			if got := obj.(*kubeovnv1.QoSPolicy).Spec.Shared; got != tc.value {
				t.Errorf("shared = %v, want %v", got, tc.value)
			}
		})
	}
}
