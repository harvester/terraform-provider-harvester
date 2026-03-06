package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceKubeOVNOvnEipStateGetter(t *testing.T) {
	testcases := []struct {
		name               string
		eip                *kubeovnv1.OvnEip
		expectedID         string
		expectedState      string
		expectedSubnet     string
		expectedV4IP       string
		expectedType       string
		expectedReady      bool
		expectedStatusV4IP string
		expectedStatusNat  string
	}{
		{
			name: "ovn eip with all fields",
			eip: &kubeovnv1.OvnEip{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ovn-eip"},
				Spec: kubeovnv1.OvnEipSpec{
					ExternalSubnet: "external-subnet",
					V4Ip:           "10.0.0.100",
					V6Ip:           "fd00::100",
					MacAddress:     "00:11:22:33:44:55",
					Type:           "nat",
				},
				Status: kubeovnv1.OvnEipStatus{
					Ready:      true,
					V4Ip:       "10.0.0.100",
					V6Ip:       "fd00::100",
					MacAddress: "00:11:22:33:44:55",
					Nat:        "test-fip",
					Type:       "nat",
				},
			},
			expectedID:         helper.BuildID("", "test-ovn-eip"),
			expectedState:      constants.StateCommonReady,
			expectedSubnet:     "external-subnet",
			expectedV4IP:       "10.0.0.100",
			expectedType:       "nat",
			expectedReady:      true,
			expectedStatusV4IP: "10.0.0.100",
			expectedStatusNat:  "test-fip",
		},
		{
			name: "empty ovn eip",
			eip: &kubeovnv1.OvnEip{
				ObjectMeta: metav1.ObjectMeta{Name: "empty-ovn-eip"},
			},
			expectedID:         helper.BuildID("", "empty-ovn-eip"),
			expectedState:      constants.StateCommonActive,
			expectedSubnet:     "",
			expectedV4IP:       "",
			expectedType:       "",
			expectedReady:      false,
			expectedStatusV4IP: "",
			expectedStatusNat:  "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNOvnEipStateGetter(tc.eip)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNOvnEip {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeKubeOVNOvnEip, getter.ResourceType)
			}
			state := getter.States[constants.FieldCommonState].(string)
			if state != tc.expectedState {
				t.Errorf("State: expected %q, got %q", tc.expectedState, state)
			}
			subnet := getter.States[constants.FieldKubeOVNOvnEipExternalSubnet].(string)
			if subnet != tc.expectedSubnet {
				t.Errorf("ExternalSubnet: expected %q, got %q", tc.expectedSubnet, subnet)
			}
			v4ip := getter.States[constants.FieldKubeOVNOvnEipV4IP].(string)
			if v4ip != tc.expectedV4IP {
				t.Errorf("V4IP: expected %q, got %q", tc.expectedV4IP, v4ip)
			}
			eipType := getter.States[constants.FieldKubeOVNOvnEipType].(string)
			if eipType != tc.expectedType {
				t.Errorf("Type: expected %q, got %q", tc.expectedType, eipType)
			}
			ready := getter.States[constants.FieldKubeOVNOvnEipStatusReady].(bool)
			if ready != tc.expectedReady {
				t.Errorf("Ready: expected %v, got %v", tc.expectedReady, ready)
			}
			statusV4IP := getter.States[constants.FieldKubeOVNOvnEipStatusV4IP].(string)
			if statusV4IP != tc.expectedStatusV4IP {
				t.Errorf("StatusV4IP: expected %q, got %q", tc.expectedStatusV4IP, statusV4IP)
			}
			statusNat := getter.States[constants.FieldKubeOVNOvnEipStatusNat].(string)
			if statusNat != tc.expectedStatusNat {
				t.Errorf("StatusNat: expected %q, got %q", tc.expectedStatusNat, statusNat)
			}
		})
	}
}
