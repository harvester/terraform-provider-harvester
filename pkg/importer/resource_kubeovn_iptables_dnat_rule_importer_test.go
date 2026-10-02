package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceKubeOVNIptablesDnatRuleStateGetter(t *testing.T) {
	testcases := []struct {
		name          string
		dnat          *kubeovnv1.IptablesDnatRule
		expectedID    string
		expectedState string
		expectedEIP   string
		expectedProto string
	}{
		{
			name: "dnat rule with all fields",
			dnat: &kubeovnv1.IptablesDnatRule{
				ObjectMeta: metav1.ObjectMeta{Name: "test-dnat"},
				Spec: kubeovnv1.IptablesDnatRuleSpec{
					EIP:          "test-eip",
					ExternalPort: "8080",
					Protocol:     "tcp",
					InternalIP:   "10.0.0.100",
					InternalPort: "80",
				},
				Status: kubeovnv1.IptablesDnatRuleStatus{
					Ready:   true,
					V4ip:    "192.168.1.100",
					NatGwDp: "test-gw",
				},
			},
			expectedID:    helper.BuildID("", "test-dnat"),
			expectedState: constants.StateCommonReady,
			expectedEIP:   "test-eip",
			expectedProto: "tcp",
		},
		{
			name: "empty dnat rule",
			dnat: &kubeovnv1.IptablesDnatRule{
				ObjectMeta: metav1.ObjectMeta{Name: "empty-dnat"},
			},
			expectedID:    helper.BuildID("", "empty-dnat"),
			expectedState: constants.StateCommonActive,
			expectedEIP:   "",
			expectedProto: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNIptablesDnatRuleStateGetter(tc.dnat)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNIptablesDnatRule {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeKubeOVNIptablesDnatRule, getter.ResourceType)
			}
			state := getter.States[constants.FieldCommonState].(string)
			if state != tc.expectedState {
				t.Errorf("State: expected %q, got %q", tc.expectedState, state)
			}
			eip := getter.States[constants.FieldKubeOVNIptablesDnatEIP].(string)
			if eip != tc.expectedEIP {
				t.Errorf("EIP: expected %q, got %q", tc.expectedEIP, eip)
			}
			proto := getter.States[constants.FieldKubeOVNIptablesDnatProtocol].(string)
			if proto != tc.expectedProto {
				t.Errorf("Protocol: expected %q, got %q", tc.expectedProto, proto)
			}
		})
	}
}
