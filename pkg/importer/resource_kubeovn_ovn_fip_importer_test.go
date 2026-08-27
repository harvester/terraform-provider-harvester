package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceKubeOVNOvnFipStateGetter(t *testing.T) {
	testcases := []struct {
		name           string
		fip            *kubeovnv1.OvnFip
		expectedID     string
		expectedState  string
		expectedOvnEip string
		expectedVpc    string
	}{
		{
			name: "ovn fip with all fields",
			fip: &kubeovnv1.OvnFip{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ovn-fip"},
				Spec: kubeovnv1.OvnFipSpec{
					OvnEip: "test-ovn-eip",
					IPType: "vip",
					IPName: "test-vip",
					Vpc:    "test-vpc",
					V4Ip:   "10.0.0.100",
					V6Ip:   "fd00::100",
				},
				Status: kubeovnv1.OvnFipStatus{
					Ready: true,
					Vpc:   "test-vpc",
					V4Eip: "192.168.1.100",
					V6Eip: "fd01::100",
					V4Ip:  "10.0.0.100",
					V6Ip:  "fd00::100",
				},
			},
			expectedID:     helper.BuildID("", "test-ovn-fip"),
			expectedState:  constants.StateCommonReady,
			expectedOvnEip: "test-ovn-eip",
			expectedVpc:    "test-vpc",
		},
		{
			name: "empty ovn fip",
			fip: &kubeovnv1.OvnFip{
				ObjectMeta: metav1.ObjectMeta{Name: "empty-ovn-fip"},
			},
			expectedID:     helper.BuildID("", "empty-ovn-fip"),
			expectedState:  constants.StateCommonActive,
			expectedOvnEip: "",
			expectedVpc:    "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNOvnFipStateGetter(tc.fip)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNOvnFip {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeKubeOVNOvnFip, getter.ResourceType)
			}
			state := getter.States[constants.FieldCommonState].(string)
			if state != tc.expectedState {
				t.Errorf("State: expected %q, got %q", tc.expectedState, state)
			}
			ovnEip := getter.States[constants.FieldKubeOVNOvnFipOvnEip].(string)
			if ovnEip != tc.expectedOvnEip {
				t.Errorf("OvnEip: expected %q, got %q", tc.expectedOvnEip, ovnEip)
			}
			vpc := getter.States[constants.FieldKubeOVNOvnFipVpc].(string)
			if vpc != tc.expectedVpc {
				t.Errorf("Vpc: expected %q, got %q", tc.expectedVpc, vpc)
			}
		})
	}
}
