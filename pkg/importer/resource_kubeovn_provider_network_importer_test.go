package importer

import (
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
		expectedCI        int
		expectedExclude   int
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
					DefaultInterface: "eth0",
					CustomInterfaces: []kubeovnv1.CustomInterface{
						{Interface: "eth1", Nodes: []string{"node1", "node2"}},
					},
					ExcludeNodes:     []string{"node3"},
					ExchangeLinkName: true,
				},
				Status: kubeovnv1.ProviderNetworkStatus{
					Ready:         true,
					ReadyNodes:    []string{"node1", "node2"},
					NotReadyNodes: []string{"node3"},
					Vlans:         []string{"vlan100"},
				},
			},
			expectedID:        helper.BuildID("", "test-pn"),
			expectedInterface: "eth0",
			expectedCI:        1,
			expectedExclude:   1,
			expectedExchange:  true,
			expectedReady:     true,
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
			expectedCI:        0,
			expectedExclude:   0,
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

			ci := getter.States[constants.FieldKubeOVNProviderNetCustomInterfaces].([]map[string]interface{})
			if len(ci) != tc.expectedCI {
				t.Errorf("CustomInterfaces: expected %d, got %d", tc.expectedCI, len(ci))
			}

			var excludeLen int
			if ex := getter.States[constants.FieldKubeOVNProviderNetExcludeNodes]; ex != nil {
				excludeLen = len(ex.([]string))
			}
			if excludeLen != tc.expectedExclude {
				t.Errorf("ExcludeNodes: expected %d, got %d", tc.expectedExclude, excludeLen)
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
