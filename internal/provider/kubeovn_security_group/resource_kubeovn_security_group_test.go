package kubeovn_security_group

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
)

func TestSecurityGroupReady(t *testing.T) {
	const portGroup = "ovn.sg.test.sg"

	testcases := []struct {
		name     string
		status   kubeovnv1.SecurityGroupStatus
		expected bool
	}{
		{name: "not reconciled yet"},
		{
			name:   "ingress synced, egress not yet",
			status: kubeovnv1.SecurityGroupStatus{IngressLastSyncSuccess: true},
		},
		{
			name:   "rules synced, port group not reported yet",
			status: kubeovnv1.SecurityGroupStatus{IngressLastSyncSuccess: true, EgressLastSyncSuccess: true},
		},
		{
			name:   "egress sync failed",
			status: kubeovnv1.SecurityGroupStatus{PortGroup: portGroup, IngressLastSyncSuccess: true},
		},
		{
			name:     "ready",
			status:   kubeovnv1.SecurityGroupStatus{PortGroup: portGroup, IngressLastSyncSuccess: true, EgressLastSyncSuccess: true},
			expected: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if got := securityGroupReady(&kubeovnv1.SecurityGroup{Status: tc.status}); got != tc.expected {
				t.Errorf("ready = %v, want %v", got, tc.expected)
			}
		})
	}
}
