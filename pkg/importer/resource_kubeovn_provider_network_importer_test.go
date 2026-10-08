package importer

import (
	"slices"
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceKubeOVNProviderNetworkStateGetter(t *testing.T) {
	testcases := []struct {
		name              string
		pn                *kubeovnv1.ProviderNetwork
		expectedID        string
		expectedInterface string
		expectedCI        []string // interface names, in order
		expectedExclude   []string
		expectedExchange  bool
		expectedReady     bool
	}{
		{
			name: "provider network with custom interfaces",
			pn: &kubeovnv1.ProviderNetwork{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pn",
				},
				Spec: kubeovnv1.ProviderNetworkSpec{
					DefaultInterface: testInterface0,
					CustomInterfaces: []kubeovnv1.CustomInterface{
						{Interface: testInterface1, Nodes: []string{testNode1, testNode2}},
					},
					ExcludeNodes:     []string{testNode3},
					ExchangeLinkName: true,
				},
				Status: kubeovnv1.ProviderNetworkStatus{
					Ready:         true,
					ReadyNodes:    []string{testNode1, testNode2},
					NotReadyNodes: []string{testNode3},
					Vlans:         []string{"vlan100"},
				},
			},
			expectedID:        helper.BuildID("", "test-pn"),
			expectedInterface: testInterface0,
			expectedCI:        []string{testInterface1},
			expectedExclude:   []string{testNode3},
			expectedExchange:  true,
			expectedReady:     true,
		},
		{
			name: "several custom interfaces keep their order",
			pn: &kubeovnv1.ProviderNetwork{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ordered-pn",
				},
				Spec: kubeovnv1.ProviderNetworkSpec{
					DefaultInterface: testInterface0,
					CustomInterfaces: []kubeovnv1.CustomInterface{
						{Interface: testInterface2, Nodes: []string{testNode2}},
						{Interface: testInterface1, Nodes: []string{testNode1}},
						{Interface: "bond0", Nodes: []string{testNode3, "node4"}},
					},
					ExcludeNodes: []string{"node6", "node5"},
				},
			},
			expectedID:        helper.BuildID("", "ordered-pn"),
			expectedInterface: testInterface0,
			expectedCI:        []string{testInterface2, testInterface1, "bond0"},
			expectedExclude:   []string{"node6", "node5"},
		},
		{
			name: "empty provider network",
			pn: &kubeovnv1.ProviderNetwork{
				ObjectMeta: metav1.ObjectMeta{
					Name: "empty-pn",
				},
				Spec: kubeovnv1.ProviderNetworkSpec{},
			},
			expectedID:        helper.BuildID("", "empty-pn"),
			expectedInterface: "",
			expectedExchange:  false,
			expectedReady:     false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNProviderNetworkStateGetter(tc.pn)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNProviderNetwork {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeKubeOVNProviderNetwork, getter.ResourceType)
			}

			iface := getter.States[constants.FieldKubeOVNProviderNetDefaultInterface].(string)
			if iface != tc.expectedInterface {
				t.Errorf("DefaultInterface: expected %q, got %q", tc.expectedInterface, iface)
			}

			ci := blockStrings(getter.States[constants.FieldKubeOVNProviderNetCustomInterfaces], constants.FieldKubeOVNCustomInterfaceInterface)
			if !slices.Equal(ci, tc.expectedCI) {
				t.Errorf("CustomInterfaces: expected %v, got %v", tc.expectedCI, ci)
			}

			exclude := getter.States[constants.FieldKubeOVNProviderNetExcludeNodes].([]string)
			if !slices.Equal(exclude, tc.expectedExclude) {
				t.Errorf("ExcludeNodes: expected %v, got %v", tc.expectedExclude, exclude)
			}

			exchange := getter.States[constants.FieldKubeOVNProviderNetExchangeLinkName].(bool)
			if exchange != tc.expectedExchange {
				t.Errorf("ExchangeLinkName: expected %v, got %v", tc.expectedExchange, exchange)
			}

			ready := getter.States[constants.FieldKubeOVNProviderNetStatusReady].(bool)
			if ready != tc.expectedReady {
				t.Errorf("StatusReady: expected %v, got %v", tc.expectedReady, ready)
			}
		})
	}
}
