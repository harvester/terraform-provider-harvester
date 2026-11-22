package kubeovn_vpc

import (
	"context"
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	kubeovnfake "github.com/kubeovn/kube-ovn/pkg/client/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func TestVpcRefresh(t *testing.T) {
	const vpcName = "test-vpc"

	testcases := []struct {
		name          string
		vpc           *kubeovnv1.Vpc // nil when the VPC is gone
		expectedState string
	}{
		{
			name:          "vpc deleted",
			expectedState: constants.StateCommonRemoved,
		},
		{
			name:          "router not created yet",
			vpc:           &kubeovnv1.Vpc{ObjectMeta: metav1.ObjectMeta{Name: vpcName}},
			expectedState: constants.StateCommonActive,
		},
		{
			name:          "router created",
			vpc:           &kubeovnv1.Vpc{ObjectMeta: metav1.ObjectMeta{Name: vpcName}, Status: kubeovnv1.VpcStatus{Standby: true}},
			expectedState: constants.StateCommonReady,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			client := kubeovnfake.NewSimpleClientset()
			if tc.vpc != nil {
				client = kubeovnfake.NewSimpleClientset(tc.vpc)
			}
			_, state, err := vpcRefresh(context.Background(), client.KubeovnV1().Vpcs(), vpcName)()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if state != tc.expectedState {
				t.Errorf("state = %q, want %q", state, tc.expectedState)
			}
		})
	}
}
