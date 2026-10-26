package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func TestResourceKubeOVNIptablesEIPStateGetter(t *testing.T) {
	const (
		macAddress = "00:00:00:12:34:56"
		natUsage   = "snat,dnat"
	)

	testcases := []struct {
		name     string
		eip      *kubeovnv1.IptablesEIP
		expected map[string]interface{}
	}{
		{
			name: "ready eip with all fields",
			eip: &kubeovnv1.IptablesEIP{
				ObjectMeta: testNatObjectMeta(testNatEIPName),
				Spec: kubeovnv1.IptablesEIPSpec{
					V4ip:           testNatExternalIP,
					V6ip:           testNatExternalIPv6,
					MacAddress:     macAddress,
					NatGwDp:        testNatGatewayName,
					QoSPolicy:      testNatQoSPolicy,
					ExternalSubnet: testNatExternalSubnet,
				},
				Status: kubeovnv1.IptablesEIPStatus{
					Ready:     true,
					IP:        testNatExternalIP,
					Nat:       natUsage,
					QoSPolicy: testNatQoSPolicy,
				},
			},
			expected: withStates(testNatMetadataStates(testNatEIPName, true), map[string]interface{}{
				constants.FieldCommonState:                      constants.StateCommonReady,
				constants.FieldKubeOVNIptablesEIPV4IP:           testNatExternalIP,
				constants.FieldKubeOVNIptablesEIPV6IP:           testNatExternalIPv6,
				constants.FieldKubeOVNIptablesEIPMacAddress:     macAddress,
				constants.FieldKubeOVNIptablesEIPNatGwDp:        testNatGatewayName,
				constants.FieldKubeOVNIptablesEIPQoSPolicy:      testNatQoSPolicy,
				constants.FieldKubeOVNIptablesEIPExternalSubnet: testNatExternalSubnet,
				constants.FieldKubeOVNIptablesEIPStatusIP:       testNatExternalIP,
				constants.FieldKubeOVNIptablesEIPStatusNat:      natUsage,
				constants.FieldKubeOVNIptablesEIPStatusQoS:      testNatQoSPolicy,
			}),
		},
		{
			name: "eip not ready yet",
			eip: &kubeovnv1.IptablesEIP{
				ObjectMeta: metav1.ObjectMeta{Name: testNatEIPName},
				Spec:       kubeovnv1.IptablesEIPSpec{NatGwDp: testNatGatewayName},
			},
			expected: withStates(testNatMetadataStates(testNatEIPName, false), map[string]interface{}{
				constants.FieldCommonState:                      constants.StateCommonActive,
				constants.FieldKubeOVNIptablesEIPV4IP:           "",
				constants.FieldKubeOVNIptablesEIPV6IP:           "",
				constants.FieldKubeOVNIptablesEIPMacAddress:     "",
				constants.FieldKubeOVNIptablesEIPNatGwDp:        testNatGatewayName,
				constants.FieldKubeOVNIptablesEIPQoSPolicy:      "",
				constants.FieldKubeOVNIptablesEIPExternalSubnet: "",
				constants.FieldKubeOVNIptablesEIPStatusIP:       "",
				constants.FieldKubeOVNIptablesEIPStatusNat:      "",
				constants.FieldKubeOVNIptablesEIPStatusQoS:      "",
			}),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNIptablesEIPStateGetter(tc.eip)
			checkNatStateGetter(t, getter, err, testNatEIPName, constants.ResourceTypeKubeOVNIptablesEIP, tc.expected)
		})
	}
}
