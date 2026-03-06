package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceKubeOVNOvnSnatRuleStateGetter(t *testing.T) {
	testcases := []struct {
		name                string
		snatRule            *kubeovnv1.OvnSnatRule
		expectedID          string
		expectedState       string
		expectedOvnEip      string
		expectedVpcSubnet   string
		expectedIPName      string
		expectedReady       bool
		expectedStatusV4Eip string
		expectedStatusVpc   string
	}{
		{
			name: "ovn snat rule with all fields",
			snatRule: &kubeovnv1.OvnSnatRule{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ovn-snat-rule"},
				Spec: kubeovnv1.OvnSnatRuleSpec{
					OvnEip:    "test-eip",
					VpcSubnet: "test-subnet",
					IPName:    "test-ip",
					Vpc:       "test-vpc",
					V4IpCidr:  "10.0.0.0/24",
					V6IpCidr:  "fd00::/64",
				},
				Status: kubeovnv1.OvnSnatRuleStatus{
					Ready: true,
					V4Eip: "10.0.0.100",
					V6Eip: "fd00::100",
					Vpc:   "test-vpc",
				},
			},
			expectedID:          helper.BuildID("", "test-ovn-snat-rule"),
			expectedState:       constants.StateCommonReady,
			expectedOvnEip:      "test-eip",
			expectedVpcSubnet:   "test-subnet",
			expectedIPName:      "test-ip",
			expectedReady:       true,
			expectedStatusV4Eip: "10.0.0.100",
			expectedStatusVpc:   "test-vpc",
		},
		{
			name: "empty ovn snat rule",
			snatRule: &kubeovnv1.OvnSnatRule{
				ObjectMeta: metav1.ObjectMeta{Name: "empty-ovn-snat-rule"},
			},
			expectedID:          helper.BuildID("", "empty-ovn-snat-rule"),
			expectedState:       constants.StateCommonActive,
			expectedOvnEip:      "",
			expectedVpcSubnet:   "",
			expectedIPName:      "",
			expectedReady:       false,
			expectedStatusV4Eip: "",
			expectedStatusVpc:   "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNOvnSnatRuleStateGetter(tc.snatRule)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNOvnSnatRule {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeKubeOVNOvnSnatRule, getter.ResourceType)
			}
			state := getter.States[constants.FieldCommonState].(string)
			if state != tc.expectedState {
				t.Errorf("State: expected %q, got %q", tc.expectedState, state)
			}
			ovnEip := getter.States[constants.FieldKubeOVNOvnSnatOvnEip].(string)
			if ovnEip != tc.expectedOvnEip {
				t.Errorf("OvnEip: expected %q, got %q", tc.expectedOvnEip, ovnEip)
			}
			vpcSubnet := getter.States[constants.FieldKubeOVNOvnSnatVpcSubnet].(string)
			if vpcSubnet != tc.expectedVpcSubnet {
				t.Errorf("VpcSubnet: expected %q, got %q", tc.expectedVpcSubnet, vpcSubnet)
			}
			ipName := getter.States[constants.FieldKubeOVNOvnSnatIPName].(string)
			if ipName != tc.expectedIPName {
				t.Errorf("IPName: expected %q, got %q", tc.expectedIPName, ipName)
			}
			ready := getter.States[constants.FieldKubeOVNOvnSnatStatusReady].(bool)
			if ready != tc.expectedReady {
				t.Errorf("Ready: expected %v, got %v", tc.expectedReady, ready)
			}
			statusV4Eip := getter.States[constants.FieldKubeOVNOvnSnatStatusV4Eip].(string)
			if statusV4Eip != tc.expectedStatusV4Eip {
				t.Errorf("StatusV4Eip: expected %q, got %q", tc.expectedStatusV4Eip, statusV4Eip)
			}
			statusVpc := getter.States[constants.FieldKubeOVNOvnSnatStatusVpc].(string)
			if statusVpc != tc.expectedStatusVpc {
				t.Errorf("StatusVpc: expected %q, got %q", tc.expectedStatusVpc, statusVpc)
			}
		})
	}
}
