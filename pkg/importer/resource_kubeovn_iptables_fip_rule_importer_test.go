package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceKubeOVNIptablesFIPRuleStateGetter(t *testing.T) {
	testcases := []struct {
		name          string
		fip           *kubeovnv1.IptablesFIPRule
		expectedID    string
		expectedState string
		expectedEIP   string
		expectedIntIP string
	}{
		{
			name: "fip rule with all fields",
			fip: &kubeovnv1.IptablesFIPRule{
				ObjectMeta: metav1.ObjectMeta{Name: "test-fip"},
				Spec: kubeovnv1.IptablesFIPRuleSpec{
					EIP:        "test-eip",
					InternalIP: "10.0.0.100",
				},
				Status: kubeovnv1.IptablesFIPRuleStatus{
					Ready:      true,
					V4ip:       "192.168.1.100",
					NatGwDp:    "test-gw",
					InternalIP: "10.0.0.100",
				},
			},
			expectedID:    helper.BuildID("", "test-fip"),
			expectedState: constants.StateCommonReady,
			expectedEIP:   "test-eip",
			expectedIntIP: "10.0.0.100",
		},
		{
			name: "empty fip rule",
			fip: &kubeovnv1.IptablesFIPRule{
				ObjectMeta: metav1.ObjectMeta{Name: "empty-fip"},
			},
			expectedID:    helper.BuildID("", "empty-fip"),
			expectedState: constants.StateCommonActive,
			expectedEIP:   "",
			expectedIntIP: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNIptablesFIPRuleStateGetter(tc.fip)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNIptablesFIPRule {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeKubeOVNIptablesFIPRule, getter.ResourceType)
			}
			state := getter.States[constants.FieldCommonState].(string)
			if state != tc.expectedState {
				t.Errorf("State: expected %q, got %q", tc.expectedState, state)
			}
			eip := getter.States[constants.FieldKubeOVNIptablesFIPEIP].(string)
			if eip != tc.expectedEIP {
				t.Errorf("EIP: expected %q, got %q", tc.expectedEIP, eip)
			}
			intIP := getter.States[constants.FieldKubeOVNIptablesFIPInternalIP].(string)
			if intIP != tc.expectedIntIP {
				t.Errorf("InternalIP: expected %q, got %q", tc.expectedIntIP, intIP)
			}
		})
	}
}
