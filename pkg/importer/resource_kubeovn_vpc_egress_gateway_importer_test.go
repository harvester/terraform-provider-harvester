package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceKubeOVNVpcEgressGatewayStateGetter(t *testing.T) {
	testcases := []struct {
		name             string
		egw              *kubeovnv1.VpcEgressGateway
		expectedID       string
		expectedState    string
		expectedVpc      string
		expectedReplicas int
		expectedExtSub   string
		expectedReady    bool
		expectedPhase    string
	}{
		{
			name: "egress gateway with all fields",
			egw: &kubeovnv1.VpcEgressGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-egw",
					Namespace: "default",
				},
				Spec: kubeovnv1.VpcEgressGatewaySpec{
					VPC:            "test-vpc",
					Replicas:       2,
					ExternalSubnet: "ext-subnet",
					InternalSubnet: "int-subnet",
					TrafficPolicy:  "Local",
					BFD: kubeovnv1.VpcEgressGatewayBFDConfig{
						Enabled:    true,
						MinRX:      100,
						MinTX:      100,
						Multiplier: 3,
					},
					Policies: []kubeovnv1.VpcEgressGatewayPolicy{
						{
							SNAT:     true,
							IPBlocks: []string{"10.0.0.0/8"},
						},
					},
				},
				Status: kubeovnv1.VpcEgressGatewayStatus{
					Ready: true,
					Phase: "Completed",
				},
			},
			expectedID:       helper.BuildID("default", "test-egw"),
			expectedState:    constants.StateCommonReady,
			expectedVpc:      "test-vpc",
			expectedReplicas: 2,
			expectedExtSub:   "ext-subnet",
			expectedReady:    true,
			expectedPhase:    "Completed",
		},
		{
			name: "empty egress gateway",
			egw: &kubeovnv1.VpcEgressGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "empty-egw",
					Namespace: "kube-system",
				},
			},
			expectedID:       helper.BuildID("kube-system", "empty-egw"),
			expectedState:    constants.StateCommonActive,
			expectedVpc:      "",
			expectedReplicas: 0,
			expectedExtSub:   "",
			expectedReady:    false,
			expectedPhase:    "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNVpcEgressGatewayStateGetter(tc.egw)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNVpcEgressGateway {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeKubeOVNVpcEgressGateway, getter.ResourceType)
			}
			state := getter.States[constants.FieldCommonState].(string)
			if state != tc.expectedState {
				t.Errorf("State: expected %q, got %q", tc.expectedState, state)
			}
			vpc := getter.States[constants.FieldKubeOVNVpcEgressGatewayVpc].(string)
			if vpc != tc.expectedVpc {
				t.Errorf("VPC: expected %q, got %q", tc.expectedVpc, vpc)
			}
			replicas := getter.States[constants.FieldKubeOVNVpcEgressGatewayReplicas].(int)
			if replicas != tc.expectedReplicas {
				t.Errorf("Replicas: expected %d, got %d", tc.expectedReplicas, replicas)
			}
			extSub := getter.States[constants.FieldKubeOVNVpcEgressGatewayExternalSubnet].(string)
			if extSub != tc.expectedExtSub {
				t.Errorf("ExternalSubnet: expected %q, got %q", tc.expectedExtSub, extSub)
			}
			ready := getter.States[constants.FieldKubeOVNVpcEgressGatewayStatusReady].(bool)
			if ready != tc.expectedReady {
				t.Errorf("Ready: expected %v, got %v", tc.expectedReady, ready)
			}
			phase := getter.States[constants.FieldKubeOVNVpcEgressGatewayStatusPhase].(string)
			if phase != tc.expectedPhase {
				t.Errorf("Phase: expected %q, got %q", tc.expectedPhase, phase)
			}
		})
	}
}
