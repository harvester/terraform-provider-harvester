package importer

import (
	"slices"
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceKubeOVNQoSPolicyStateGetter(t *testing.T) {
	testcases := []struct {
		name           string
		qos            *kubeovnv1.QoSPolicy
		expectedID     string
		expectedShared bool
		expectedBT     string
		expectedRules  []string // names, in order
		expectedRates  []string // rate_max values, in order
	}{
		{
			name: "qos policy with bandwidth rules",
			qos: &kubeovnv1.QoSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: testQoSName,
				},
				Spec: kubeovnv1.QoSPolicySpec{
					Shared:      true,
					BindingType: kubeovnv1.QoSBindingTypeEIP,
					BandwidthLimitRules: kubeovnv1.QoSPolicyBandwidthLimitRules{
						{
							Name:      "rule1",
							RateMax:   "100M",
							BurstMax:  "200M",
							Direction: "ingress",
						},
					},
				},
				Status: kubeovnv1.QoSPolicyStatus{
					Shared:      true,
					BindingType: kubeovnv1.QoSBindingTypeEIP,
				},
			},
			expectedID:     helper.BuildID("", testQoSName),
			expectedShared: true,
			expectedBT:     string(kubeovnv1.QoSBindingTypeEIP),
			expectedRules:  []string{"rule1"},
			expectedRates:  []string{"100M"},
		},
		{
			name: "several bandwidth rules keep their order",
			qos: &kubeovnv1.QoSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ordered-qos",
				},
				Spec: kubeovnv1.QoSPolicySpec{
					BindingType: kubeovnv1.QoSBindingTypeNatGw,
					BandwidthLimitRules: kubeovnv1.QoSPolicyBandwidthLimitRules{
						{Name: "egress-limit", RateMax: "10M", Direction: kubeovnv1.QoSDirectionEgress},
						{Name: "ingress-limit", RateMax: "20M", Direction: kubeovnv1.QoSDirectionIngress},
						{Name: "ingress-burst", RateMax: "30M", Direction: kubeovnv1.QoSDirectionIngress},
					},
				},
			},
			expectedID:    helper.BuildID("", "ordered-qos"),
			expectedBT:    string(kubeovnv1.QoSBindingTypeNatGw),
			expectedRules: []string{"egress-limit", "ingress-limit", "ingress-burst"},
			expectedRates: []string{"10M", "20M", "30M"},
		},
		{
			name: "empty qos policy",
			qos: &kubeovnv1.QoSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "empty-qos",
				},
				Spec: kubeovnv1.QoSPolicySpec{
					BindingType: kubeovnv1.QoSBindingTypeNatGw,
				},
			},
			expectedID:     helper.BuildID("", "empty-qos"),
			expectedShared: false,
			expectedBT:     string(kubeovnv1.QoSBindingTypeNatGw),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNQoSPolicyStateGetter(tc.qos)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNQoSPolicy {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeKubeOVNQoSPolicy, getter.ResourceType)
			}

			shared := getter.States[constants.FieldKubeOVNQoSShared].(bool)
			if shared != tc.expectedShared {
				t.Errorf("Shared: expected %v, got %v", tc.expectedShared, shared)
			}

			bt := getter.States[constants.FieldKubeOVNQoSBindingType].(string)
			if bt != tc.expectedBT {
				t.Errorf("BindingType: expected %q, got %q", tc.expectedBT, bt)
			}

			rules := blockStrings(getter.States[constants.FieldKubeOVNQoSBandwidthLimitRules], constants.FieldKubeOVNQoSRuleName)
			if !slices.Equal(rules, tc.expectedRules) {
				t.Errorf("BandwidthLimitRules: expected %v, got %v", tc.expectedRules, rules)
			}

			rates := blockStrings(getter.States[constants.FieldKubeOVNQoSBandwidthLimitRules], constants.FieldKubeOVNQoSRuleRateMax)
			if !slices.Equal(rates, tc.expectedRates) {
				t.Errorf("Rule RateMax: expected %v, got %v", tc.expectedRates, rates)
			}
		})
	}
}
