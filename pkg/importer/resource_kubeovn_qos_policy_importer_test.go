package importer

import (
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
		expectedRules  int
	}{
		{
			name: "qos policy with bandwidth rules",
			qos: &kubeovnv1.QoSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-qos",
				},
				Spec: kubeovnv1.QoSPolicySpec{
					Shared:      true,
					BindingType: "EIP",
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
					BindingType: "EIP",
				},
			},
			expectedID:     helper.BuildID("", "test-qos"),
			expectedShared: true,
			expectedBT:     "EIP",
			expectedRules:  1,
		},
		{
			name: "empty qos policy",
			qos: &kubeovnv1.QoSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "empty-qos",
				},
				Spec: kubeovnv1.QoSPolicySpec{
					BindingType: "NATGW",
				},
			},
			expectedID:     helper.BuildID("", "empty-qos"),
			expectedShared: false,
			expectedBT:     "NATGW",
			expectedRules:  0,
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

			rules := getter.States[constants.FieldKubeOVNQoSBandwidthLimitRules].([]map[string]interface{})
			if len(rules) != tc.expectedRules {
				t.Errorf("BandwidthLimitRules: expected %d, got %d", tc.expectedRules, len(rules))
			}

			if tc.expectedRules > 0 {
				rule := rules[0]
				if rule[constants.FieldKubeOVNQoSRuleName] != "rule1" {
					t.Errorf("Rule Name: expected %q, got %q", "rule1", rule[constants.FieldKubeOVNQoSRuleName])
				}
				if rule[constants.FieldKubeOVNQoSRuleRateMax] != "100M" {
					t.Errorf("Rule RateMax: expected %q, got %q", "100M", rule[constants.FieldKubeOVNQoSRuleRateMax])
				}
			}
		})
	}
}
