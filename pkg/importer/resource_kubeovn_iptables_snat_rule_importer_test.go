package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func TestResourceKubeOVNIptablesSnatRuleStateGetter(t *testing.T) {
	const snatName = "test-snat"

	testcases := []struct {
		name     string
		snat     *kubeovnv1.IptablesSnatRule
		expected map[string]interface{}
	}{
		{
			name: "ready snat rule with all fields",
			snat: &kubeovnv1.IptablesSnatRule{
				ObjectMeta: testNatObjectMeta(snatName),
				Spec: kubeovnv1.IptablesSnatRuleSpec{
					EIP:          testNatEIPName,
					InternalCIDR: testCIDR,
				},
				Status: kubeovnv1.IptablesSnatRuleStatus{
					Ready:        true,
					V4ip:         testNatExternalIP,
					V6ip:         testNatExternalIPv6,
					NatGwDp:      testNatGatewayName,
					InternalCIDR: testCIDR,
				},
			},
			expected: withStates(testNatMetadataStates(snatName, true), map[string]interface{}{
				constants.FieldCommonState:                     constants.StateCommonReady,
				constants.FieldKubeOVNIptablesSnatEIP:          testNatEIPName,
				constants.FieldKubeOVNIptablesSnatInternalCIDR: testCIDR,
				constants.FieldKubeOVNIptablesSnatStatusV4IP:   testNatExternalIP,
				constants.FieldKubeOVNIptablesSnatStatusV6IP:   testNatExternalIPv6,
				constants.FieldKubeOVNIptablesSnatStatusNat:    testNatGatewayName,
				constants.FieldKubeOVNIptablesSnatStatusCIDR:   testCIDR,
			}),
		},
		{
			name: "snat rule not ready yet",
			snat: &kubeovnv1.IptablesSnatRule{
				ObjectMeta: metav1.ObjectMeta{Name: snatName},
				Spec: kubeovnv1.IptablesSnatRuleSpec{
					EIP:          testNatEIPName,
					InternalCIDR: testCIDR,
				},
			},
			expected: withStates(testNatMetadataStates(snatName, false), map[string]interface{}{
				constants.FieldCommonState:                     constants.StateCommonActive,
				constants.FieldKubeOVNIptablesSnatEIP:          testNatEIPName,
				constants.FieldKubeOVNIptablesSnatInternalCIDR: testCIDR,
				constants.FieldKubeOVNIptablesSnatStatusV4IP:   "",
				constants.FieldKubeOVNIptablesSnatStatusV6IP:   "",
				constants.FieldKubeOVNIptablesSnatStatusNat:    "",
				constants.FieldKubeOVNIptablesSnatStatusCIDR:   "",
			}),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNIptablesSnatRuleStateGetter(tc.snat)
			checkNatStateGetter(t, getter, err, snatName, constants.ResourceTypeKubeOVNIptablesSnatRule, tc.expected)
		})
	}
}
