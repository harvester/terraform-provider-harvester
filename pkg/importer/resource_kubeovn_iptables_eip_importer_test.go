package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceKubeOVNIptablesEIPStateGetter(t *testing.T) {
	testcases := []struct {
		name          string
		eip           *kubeovnv1.IptablesEIP
		expectedID    string
		expectedState string
		expectedNatGw string
		expectedV4IP  string
		expectedReady bool
	}{
		{
			name: "eip with all fields",
			eip: &kubeovnv1.IptablesEIP{
				ObjectMeta: metav1.ObjectMeta{Name: "test-eip"},
				Spec: kubeovnv1.IptablesEIPSpec{
					V4ip:           "192.168.1.100",
					NatGwDp:        "test-gw",
					ExternalSubnet: "external-subnet",
				},
				Status: kubeovnv1.IptablesEIPStatus{
					Ready: true,
					IP:    "192.168.1.100",
					Nat:   "test-gw",
				},
			},
			expectedID:    helper.BuildID("", "test-eip"),
			expectedState: constants.StateCommonReady,
			expectedNatGw: "test-gw",
			expectedV4IP:  "192.168.1.100",
			expectedReady: true,
		},
		{
			name: "empty eip",
			eip: &kubeovnv1.IptablesEIP{
				ObjectMeta: metav1.ObjectMeta{Name: "empty-eip"},
			},
			expectedID:    helper.BuildID("", "empty-eip"),
			expectedState: constants.StateCommonActive,
			expectedNatGw: "",
			expectedV4IP:  "",
			expectedReady: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNIptablesEIPStateGetter(tc.eip)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNIptablesEIP {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeKubeOVNIptablesEIP, getter.ResourceType)
			}
			state := getter.States[constants.FieldCommonState].(string)
			if state != tc.expectedState {
				t.Errorf("State: expected %q, got %q", tc.expectedState, state)
			}
			natGw := getter.States[constants.FieldKubeOVNIptablesEIPNatGwDp].(string)
			if natGw != tc.expectedNatGw {
				t.Errorf("NatGwDp: expected %q, got %q", tc.expectedNatGw, natGw)
			}
			v4ip := getter.States[constants.FieldKubeOVNIptablesEIPV4IP].(string)
			if v4ip != tc.expectedV4IP {
				t.Errorf("V4IP: expected %q, got %q", tc.expectedV4IP, v4ip)
			}
			ready := getter.States[constants.FieldKubeOVNIptablesEIPReady].(bool)
			if ready != tc.expectedReady {
				t.Errorf("Ready: expected %v, got %v", tc.expectedReady, ready)
			}
		})
	}
}
