package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceKubeOVNVpcNatGatewayStateGetter(t *testing.T) {
	testcases := []struct {
		name            string
		gw              *kubeovnv1.VpcNatGateway
		expectedID      string
		expectedVpc     string
		expectedSubnet  string
		expectedLanIP   string
		expectedExtSubs int
	}{
		{
			name: "gateway with all fields",
			gw: &kubeovnv1.VpcNatGateway{
				ObjectMeta: metav1.ObjectMeta{Name: "test-gw"},
				Spec: kubeovnv1.VpcNatSpec{
					Vpc:             "test-vpc",
					Subnet:          "test-subnet",
					LanIP:           "10.0.0.100",
					ExternalSubnets: []string{"ext-sub-1", "ext-sub-2"},
					Selector:        []string{"kubernetes.io/os=linux"},
					QoSPolicy:       "test-qos",
				},
				Status: kubeovnv1.VpcNatStatus{
					QoSPolicy:       "test-qos",
					ExternalSubnets: []string{"ext-sub-1", "ext-sub-2"},
				},
			},
			expectedID:      helper.BuildID("", "test-gw"),
			expectedVpc:     "test-vpc",
			expectedSubnet:  "test-subnet",
			expectedLanIP:   "10.0.0.100",
			expectedExtSubs: 2,
		},
		{
			name: "empty gateway",
			gw: &kubeovnv1.VpcNatGateway{
				ObjectMeta: metav1.ObjectMeta{Name: "empty-gw"},
			},
			expectedID:      helper.BuildID("", "empty-gw"),
			expectedVpc:     "",
			expectedSubnet:  "",
			expectedLanIP:   "",
			expectedExtSubs: 0,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNVpcNatGatewayStateGetter(tc.gw)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNVpcNatGateway {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeKubeOVNVpcNatGateway, getter.ResourceType)
			}
			state := getter.States[constants.FieldCommonState].(string)
			if state != constants.StateCommonActive {
				t.Errorf("State: expected %q, got %q", constants.StateCommonActive, state)
			}
			vpc := getter.States[constants.FieldKubeOVNVpcNatGwVpc].(string)
			if vpc != tc.expectedVpc {
				t.Errorf("Vpc: expected %q, got %q", tc.expectedVpc, vpc)
			}
			subnet := getter.States[constants.FieldKubeOVNVpcNatGwSubnet].(string)
			if subnet != tc.expectedSubnet {
				t.Errorf("Subnet: expected %q, got %q", tc.expectedSubnet, subnet)
			}
			lanIP := getter.States[constants.FieldKubeOVNVpcNatGwLanIP].(string)
			if lanIP != tc.expectedLanIP {
				t.Errorf("LanIP: expected %q, got %q", tc.expectedLanIP, lanIP)
			}
			extSubs := getter.States[constants.FieldKubeOVNVpcNatGwExternalSubnets].([]string)
			if len(extSubs) != tc.expectedExtSubs {
				t.Errorf("ExternalSubnets: expected %d items, got %d", tc.expectedExtSubs, len(extSubs))
			}
		})
	}
}
