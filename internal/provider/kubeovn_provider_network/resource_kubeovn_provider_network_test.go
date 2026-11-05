package kubeovn_provider_network

import (
	"context"
	"errors"
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	kubeovnfake "github.com/kubeovn/kube-ovn/pkg/client/clientset/versioned/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func TestProviderNetworkRefresh(t *testing.T) {
	errAPI := errors.New("api unavailable")
	readyCondition := kubeovnv1.ProviderNetworkCondition{
		Node:      node1,
		Condition: kubeovnv1.Condition{Type: kubeovnv1.Ready, Status: corev1.ConditionTrue},
	}

	testcases := []struct {
		name          string
		status        *kubeovnv1.ProviderNetworkStatus // nil when the provider network is gone
		getErr        error
		expectedState string
		expectedErr   error
	}{
		{
			name:          "provider network deleted",
			expectedState: constants.StateCommonRemoved,
		},
		{
			name:          "not reconciled yet",
			status:        &kubeovnv1.ProviderNetworkStatus{},
			expectedState: constants.StateCommonActive,
		},
		{
			// IsReady() is true without any not-ready condition, so only
			// status.ready tells that every node was checked.
			name:          "node conditions ready but status not set yet",
			status:        &kubeovnv1.ProviderNetworkStatus{Conditions: []kubeovnv1.ProviderNetworkCondition{readyCondition}},
			expectedState: constants.StateCommonActive,
		},
		{
			name:          "bridge missing on a node",
			status:        &kubeovnv1.ProviderNetworkStatus{ReadyNodes: []string{node1}, NotReadyNodes: []string{node2}},
			expectedState: constants.StateCommonActive,
		},
		{
			name:          "ready on every node",
			status:        &kubeovnv1.ProviderNetworkStatus{Ready: true, ReadyNodes: []string{node1}, Conditions: []kubeovnv1.ProviderNetworkCondition{readyCondition}},
			expectedState: constants.StateCommonReady,
		},
		{
			name:          "API error",
			status:        &kubeovnv1.ProviderNetworkStatus{},
			getErr:        errAPI,
			expectedState: constants.StateCommonError,
			expectedErr:   errAPI,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			client := kubeovnfake.NewSimpleClientset()
			providerNetworks := client.KubeovnV1().ProviderNetworks()
			if tc.status != nil {
				// Created through the client: the fake tracker would file an
				// object given to NewSimpleClientset under "providernetworks"
				// while the client reads "provider-networks".
				_, err := providerNetworks.Create(context.Background(), &kubeovnv1.ProviderNetwork{
					ObjectMeta: metav1.ObjectMeta{Name: pnName},
					Status:     *tc.status,
				}, metav1.CreateOptions{})
				if err != nil {
					t.Fatal(err)
				}
			}
			if tc.getErr != nil {
				client.PrependReactor("get", "provider-networks", func(k8stesting.Action) (bool, runtime.Object, error) {
					return true, nil, tc.getErr
				})
			}
			_, state, err := providerNetworkRefresh(context.Background(), providerNetworks, pnName)()
			if state != tc.expectedState {
				t.Errorf("state = %q, want %q", state, tc.expectedState)
			}
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("error = %v, want %v", err, tc.expectedErr)
			}
		})
	}
}
