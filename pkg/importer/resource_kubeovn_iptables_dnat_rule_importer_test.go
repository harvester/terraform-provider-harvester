package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func TestResourceKubeOVNIptablesDnatRuleStateGetter(t *testing.T) {
	const (
		dnatName     = "test-dnat"
		externalPort = "8022"
		internalPort = "22"
		protocol     = "udp"
	)

	testcases := []struct {
		name     string
		dnat     *kubeovnv1.IptablesDnatRule
		expected map[string]interface{}
	}{
		{
			name: "ready dnat rule with all fields",
			dnat: &kubeovnv1.IptablesDnatRule{
				ObjectMeta: testNatObjectMeta(dnatName),
				Spec: kubeovnv1.IptablesDnatRuleSpec{
					EIP:          testNatEIPName,
					ExternalPort: externalPort,
					Protocol:     protocol,
					InternalIP:   testNatInternalIP,
					InternalPort: internalPort,
				},
				Status: kubeovnv1.IptablesDnatRuleStatus{
					Ready:        true,
					V4ip:         testNatExternalIP,
					V6ip:         testNatExternalIPv6,
					NatGwDp:      testNatGatewayName,
					Protocol:     protocol,
					InternalIP:   testNatInternalIP,
					InternalPort: internalPort,
					ExternalPort: externalPort,
				},
			},
			expected: withStates(testNatMetadataStates(dnatName, true), map[string]interface{}{
				constants.FieldCommonState:                     constants.StateCommonReady,
				constants.FieldKubeOVNIptablesDnatEIP:          testNatEIPName,
				constants.FieldKubeOVNIptablesDnatExternalPort: externalPort,
				constants.FieldKubeOVNIptablesDnatProtocol:     protocol,
				constants.FieldKubeOVNIptablesDnatInternalIP:   testNatInternalIP,
				constants.FieldKubeOVNIptablesDnatInternalPort: internalPort,
				constants.FieldKubeOVNIptablesDnatStatusV4IP:   testNatExternalIP,
				constants.FieldKubeOVNIptablesDnatStatusV6IP:   testNatExternalIPv6,
				constants.FieldKubeOVNIptablesDnatStatusNat:    testNatGatewayName,
				constants.FieldKubeOVNIptablesDnatStatusProto:  protocol,
				constants.FieldKubeOVNIptablesDnatStatusIntIP:  testNatInternalIP,
				constants.FieldKubeOVNIptablesDnatStatusIntP:   internalPort,
				constants.FieldKubeOVNIptablesDnatStatusExtP:   externalPort,
			}),
		},
		{
			name: "dnat rule not ready yet",
			dnat: &kubeovnv1.IptablesDnatRule{
				ObjectMeta: metav1.ObjectMeta{Name: dnatName},
				Spec: kubeovnv1.IptablesDnatRuleSpec{
					EIP:          testNatEIPName,
					ExternalPort: externalPort,
					Protocol:     protocol,
					InternalIP:   testNatInternalIP,
					InternalPort: internalPort,
				},
			},
			expected: withStates(testNatMetadataStates(dnatName, false), map[string]interface{}{
				constants.FieldCommonState:                     constants.StateCommonActive,
				constants.FieldKubeOVNIptablesDnatEIP:          testNatEIPName,
				constants.FieldKubeOVNIptablesDnatExternalPort: externalPort,
				constants.FieldKubeOVNIptablesDnatProtocol:     protocol,
				constants.FieldKubeOVNIptablesDnatInternalIP:   testNatInternalIP,
				constants.FieldKubeOVNIptablesDnatInternalPort: internalPort,
				constants.FieldKubeOVNIptablesDnatStatusV4IP:   "",
				constants.FieldKubeOVNIptablesDnatStatusV6IP:   "",
				constants.FieldKubeOVNIptablesDnatStatusNat:    "",
				constants.FieldKubeOVNIptablesDnatStatusProto:  "",
				constants.FieldKubeOVNIptablesDnatStatusIntIP:  "",
				constants.FieldKubeOVNIptablesDnatStatusIntP:   "",
				constants.FieldKubeOVNIptablesDnatStatusExtP:   "",
			}),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNIptablesDnatRuleStateGetter(tc.dnat)
			checkNatStateGetter(t, getter, err, dnatName, constants.ResourceTypeKubeOVNIptablesDnatRule, tc.expected)
		})
	}
}
