package kubeovn_qos_policy

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
)

func TestQoSPolicyReady(t *testing.T) {
	testcases := []struct {
		name     string
		spec     kubeovnv1.QoSPolicyBindingType
		status   kubeovnv1.QoSPolicyBindingType
		expected bool
	}{
		{name: "not validated yet", spec: kubeovnv1.QoSBindingTypeEIP},
		{name: "validated EIP policy", spec: kubeovnv1.QoSBindingTypeEIP, status: kubeovnv1.QoSBindingTypeEIP, expected: true},
		{name: "validated NAT gateway policy", spec: kubeovnv1.QoSBindingTypeNatGw, status: kubeovnv1.QoSBindingTypeNatGw, expected: true},
		{name: "status from another binding type", spec: kubeovnv1.QoSBindingTypeNatGw, status: kubeovnv1.QoSBindingTypeEIP},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			qos := &kubeovnv1.QoSPolicy{
				Spec:   kubeovnv1.QoSPolicySpec{BindingType: tc.spec},
				Status: kubeovnv1.QoSPolicyStatus{BindingType: tc.status},
			}
			if got := qosPolicyReady(qos); got != tc.expected {
				t.Errorf("ready = %v, want %v", got, tc.expected)
			}
		})
	}
}
