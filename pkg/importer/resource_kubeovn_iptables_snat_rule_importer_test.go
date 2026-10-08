package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceKubeOVNIptablesSnatRuleStateGetter(t *testing.T) {
	testcases := []struct {
		name          string
		snat          *kubeovnv1.IptablesSnatRule
		expectedID    string
		expectedState string
		expectedEIP   string
		expectedCIDR  string
	}{
		{
			name: "snat rule with all fields",
			snat: &kubeovnv1.IptablesSnatRule{
				ObjectMeta: metav1.ObjectMeta{Name: "test-snat"},
				Spec: kubeovnv1.IptablesSnatRuleSpec{
					EIP:          testEIPName,
					InternalCIDR: testCIDR,
				},
				Status: kubeovnv1.IptablesSnatRuleStatus{
					Ready:        true,
					V4ip:         testExternalIP,
					NatGwDp:      testNatGatewayName,
					InternalCIDR: testCIDR,
				},
			},
			expectedID:    helper.BuildID("", "test-snat"),
			expectedState: constants.StateCommonReady,
			expectedEIP:   testEIPName,
			expectedCIDR:  testCIDR,
		},
		{
			name: "empty snat rule",
			snat: &kubeovnv1.IptablesSnatRule{
				ObjectMeta: metav1.ObjectMeta{Name: "empty-snat"},
			},
			expectedID:    helper.BuildID("", "empty-snat"),
			expectedState: constants.StateCommonActive,
			expectedEIP:   "",
			expectedCIDR:  "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNIptablesSnatRuleStateGetter(tc.snat)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNIptablesSnatRule {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeKubeOVNIptablesSnatRule, getter.ResourceType)
			}
			state := getter.States[constants.FieldCommonState].(string)
			if state != tc.expectedState {
				t.Errorf("State: expected %q, got %q", tc.expectedState, state)
			}
			eip := getter.States[constants.FieldKubeOVNIptablesSnatEIP].(string)
			if eip != tc.expectedEIP {
				t.Errorf("EIP: expected %q, got %q", tc.expectedEIP, eip)
			}
			cidr := getter.States[constants.FieldKubeOVNIptablesSnatInternalCIDR].(string)
			if cidr != tc.expectedCIDR {
				t.Errorf("InternalCIDR: expected %q, got %q", tc.expectedCIDR, cidr)
			}
		})
	}
}
