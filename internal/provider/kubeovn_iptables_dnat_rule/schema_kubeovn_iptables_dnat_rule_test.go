package kubeovn_iptables_dnat_rule

import (
	"testing"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// The plan must accept exactly the protocols the kube-ovn webhook accepts,
// which compares them case-insensitively.
func TestDnatProtocolValidation(t *testing.T) {
	validate := Schema()[constants.FieldKubeOVNIptablesDnatProtocol].ValidateFunc

	testcases := []struct {
		protocol    string
		expectedErr bool
	}{
		{protocol: "tcp"},
		{protocol: "udp"},
		{protocol: "TCP"},
		{protocol: "Udp"},
		{protocol: "icmp", expectedErr: true},
		{protocol: "sctp", expectedErr: true},
		{protocol: "", expectedErr: true},
	}

	for _, tc := range testcases {
		t.Run(tc.protocol, func(t *testing.T) {
			_, errs := validate(tc.protocol, constants.FieldKubeOVNIptablesDnatProtocol)
			if tc.expectedErr != (len(errs) > 0) {
				t.Errorf("errors = %v, want an error: %v", errs, tc.expectedErr)
			}
		})
	}
}
