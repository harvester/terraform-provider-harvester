package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceKubeOVNVpcDnsStateGetter(t *testing.T) {
	testcases := []struct {
		name             string
		vpcdns           *kubeovnv1.VpcDns
		expectedID       string
		expectedState    string
		expectedReplicas int
		expectedVpc      string
		expectedSubnet   string
		expectedActive   bool
	}{
		{
			name: "vpc dns with all fields",
			vpcdns: &kubeovnv1.VpcDns{
				ObjectMeta: metav1.ObjectMeta{Name: "test-vpc-dns"},
				Spec: kubeovnv1.VpcDNSSpec{
					Replicas: 3,
					Vpc:      "test-vpc",
					Subnet:   "test-subnet",
				},
				Status: kubeovnv1.VpcDNSStatus{
					Active: true,
				},
			},
			expectedID:       helper.BuildID("", "test-vpc-dns"),
			expectedState:    constants.StateCommonReady,
			expectedReplicas: 3,
			expectedVpc:      "test-vpc",
			expectedSubnet:   "test-subnet",
			expectedActive:   true,
		},
		{
			name: "empty vpc dns",
			vpcdns: &kubeovnv1.VpcDns{
				ObjectMeta: metav1.ObjectMeta{Name: "empty-vpc-dns"},
			},
			expectedID:       helper.BuildID("", "empty-vpc-dns"),
			expectedState:    constants.StateCommonActive,
			expectedReplicas: 0,
			expectedVpc:      "",
			expectedSubnet:   "",
			expectedActive:   false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNVpcDnsStateGetter(tc.vpcdns)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNVpcDns {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeKubeOVNVpcDns, getter.ResourceType)
			}
			state := getter.States[constants.FieldCommonState].(string)
			if state != tc.expectedState {
				t.Errorf("State: expected %q, got %q", tc.expectedState, state)
			}
			replicas := getter.States[constants.FieldKubeOVNVpcDnsReplicas].(int)
			if replicas != tc.expectedReplicas {
				t.Errorf("Replicas: expected %d, got %d", tc.expectedReplicas, replicas)
			}
			vpc := getter.States[constants.FieldKubeOVNVpcDnsVpc].(string)
			if vpc != tc.expectedVpc {
				t.Errorf("Vpc: expected %q, got %q", tc.expectedVpc, vpc)
			}
			subnet := getter.States[constants.FieldKubeOVNVpcDnsSubnet].(string)
			if subnet != tc.expectedSubnet {
				t.Errorf("Subnet: expected %q, got %q", tc.expectedSubnet, subnet)
			}
			active := getter.States[constants.FieldKubeOVNVpcDnsStatusActive].(bool)
			if active != tc.expectedActive {
				t.Errorf("Active: expected %v, got %v", tc.expectedActive, active)
			}
		})
	}
}
