package kubeovn_ippool

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestIPPoolReady(t *testing.T) {
	condition := func(conditionType kubeovnv1.ConditionType, status corev1.ConditionStatus) kubeovnv1.Condition {
		return kubeovnv1.Condition{Type: conditionType, Status: status}
	}

	testcases := []struct {
		name       string
		conditions []kubeovnv1.Condition
		expected   bool
	}{
		{name: "not reconciled yet"},
		{
			name:       "standard conditions only",
			conditions: []kubeovnv1.Condition{condition(kubeovnv1.Ready, corev1.ConditionUnknown), condition(kubeovnv1.Error, corev1.ConditionUnknown)},
		},
		{
			name:       "IPAM update failed",
			conditions: []kubeovnv1.Condition{condition(kubeovnv1.Ready, corev1.ConditionFalse), condition(kubeovnv1.Error, corev1.ConditionTrue)},
		},
		{
			name:       "ready",
			conditions: []kubeovnv1.Condition{condition(kubeovnv1.Ready, corev1.ConditionTrue), condition(kubeovnv1.Error, corev1.ConditionUnknown)},
			expected:   true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ippoolReady(&kubeovnv1.IPPool{Status: kubeovnv1.IPPoolStatus{Conditions: tc.conditions}}); got != tc.expected {
				t.Errorf("ready = %v, want %v", got, tc.expected)
			}
		})
	}
}
