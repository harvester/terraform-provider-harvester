package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceKubeOVNOvnDnatRuleStateGetter(t *testing.T) {
	testcases := []struct {
		name                 string
		rule                 *kubeovnv1.OvnDnatRule
		expectedID           string
		expectedState        string
		expectedOvnEip       string
		expectedInternalPort string
		expectedExternalPort string
		expectedProtocol     string
		expectedReady        bool
		expectedStatusV4Eip  string
		expectedStatusIPName string
	}{
		{
			name: "ovn dnat rule with all fields",
			rule: &kubeovnv1.OvnDnatRule{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ovn-dnat-rule"},
				Spec: kubeovnv1.OvnDnatRuleSpec{
					OvnEip:       "test-eip",
					IPType:       "v4",
					IPName:       "test-ip",
					InternalPort: "8080",
					ExternalPort: "80",
					Protocol:     "tcp",
					Vpc:          "test-vpc",
					V4Ip:         "10.0.0.100",
					V6Ip:         "fd00::100",
				},
				Status: kubeovnv1.OvnDnatRuleStatus{
					Ready:  true,
					V4Eip:  "10.0.0.1",
					V6Eip:  "fd00::1",
					Vpc:    "test-vpc",
					IPName: "test-ip",
				},
			},
			expectedID:           helper.BuildID("", "test-ovn-dnat-rule"),
			expectedState:        constants.StateCommonReady,
			expectedOvnEip:       "test-eip",
			expectedInternalPort: "8080",
			expectedExternalPort: "80",
			expectedProtocol:     "tcp",
			expectedReady:        true,
			expectedStatusV4Eip:  "10.0.0.1",
			expectedStatusIPName: "test-ip",
		},
		{
			name: "empty ovn dnat rule",
			rule: &kubeovnv1.OvnDnatRule{
				ObjectMeta: metav1.ObjectMeta{Name: "empty-ovn-dnat-rule"},
			},
			expectedID:           helper.BuildID("", "empty-ovn-dnat-rule"),
			expectedState:        constants.StateCommonActive,
			expectedOvnEip:       "",
			expectedInternalPort: "",
			expectedExternalPort: "",
			expectedProtocol:     "",
			expectedReady:        false,
			expectedStatusV4Eip:  "",
			expectedStatusIPName: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNOvnDnatRuleStateGetter(tc.rule)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNOvnDnatRule {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeKubeOVNOvnDnatRule, getter.ResourceType)
			}
			state := getter.States[constants.FieldCommonState].(string)
			if state != tc.expectedState {
				t.Errorf("State: expected %q, got %q", tc.expectedState, state)
			}
			ovnEip := getter.States[constants.FieldKubeOVNOvnDnatOvnEip].(string)
			if ovnEip != tc.expectedOvnEip {
				t.Errorf("OvnEip: expected %q, got %q", tc.expectedOvnEip, ovnEip)
			}
			internalPort := getter.States[constants.FieldKubeOVNOvnDnatInternalPort].(string)
			if internalPort != tc.expectedInternalPort {
				t.Errorf("InternalPort: expected %q, got %q", tc.expectedInternalPort, internalPort)
			}
			externalPort := getter.States[constants.FieldKubeOVNOvnDnatExternalPort].(string)
			if externalPort != tc.expectedExternalPort {
				t.Errorf("ExternalPort: expected %q, got %q", tc.expectedExternalPort, externalPort)
			}
			protocol := getter.States[constants.FieldKubeOVNOvnDnatProtocol].(string)
			if protocol != tc.expectedProtocol {
				t.Errorf("Protocol: expected %q, got %q", tc.expectedProtocol, protocol)
			}
			ready := getter.States[constants.FieldKubeOVNOvnDnatStatusReady].(bool)
			if ready != tc.expectedReady {
				t.Errorf("Ready: expected %v, got %v", tc.expectedReady, ready)
			}
			statusV4Eip := getter.States[constants.FieldKubeOVNOvnDnatStatusV4Eip].(string)
			if statusV4Eip != tc.expectedStatusV4Eip {
				t.Errorf("StatusV4Eip: expected %q, got %q", tc.expectedStatusV4Eip, statusV4Eip)
			}
			statusIPName := getter.States[constants.FieldKubeOVNOvnDnatStatusIPName].(string)
			if statusIPName != tc.expectedStatusIPName {
				t.Errorf("StatusIPName: expected %q, got %q", tc.expectedStatusIPName, statusIPName)
			}
		})
	}
}
