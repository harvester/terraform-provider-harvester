package kubeovn_provider_network

import (
	"context"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

const (
	pnName      = "test-pn"
	pnInterface = "eth1"
)

func customInterface(name string, nodes ...interface{}) map[string]interface{} {
	return map[string]interface{}{
		constants.FieldKubeOVNCustomInterfaceInterface: name,
		constants.FieldKubeOVNCustomInterfaceNodes:     nodes,
	}
}

func TestProviderNetworkConstructor(t *testing.T) {
	const (
		node1 = "node1"
		node2 = "node2"
	)

	existing := &kubeovnv1.ProviderNetwork{
		ObjectMeta: metav1.ObjectMeta{Name: pnName},
		Spec: kubeovnv1.ProviderNetworkSpec{
			DefaultInterface: "eth0",
			CustomInterfaces: []kubeovnv1.CustomInterface{{Interface: "old0", Nodes: []string{"old-node"}}},
			ExcludeNodes:     []string{"old-node"},
		},
	}

	testcases := []struct {
		name                     string
		existing                 *kubeovnv1.ProviderNetwork // nil creates the provider network
		config                   map[string]interface{}
		expectedCustomInterfaces []kubeovnv1.CustomInterface
		expectedExcludeNodes     []string
	}{
		{
			name: "create keeps the order of several entries",
			config: map[string]interface{}{
				constants.FieldKubeOVNProviderNetCustomInterfaces: []interface{}{
					customInterface("eth2", node2, node1),
					customInterface(pnInterface, node1),
				},
				constants.FieldKubeOVNProviderNetExcludeNodes: []interface{}{node2, node1},
			},
			expectedCustomInterfaces: []kubeovnv1.CustomInterface{
				{Interface: "eth2", Nodes: []string{node2, node1}},
				{Interface: pnInterface, Nodes: []string{node1}},
			},
			expectedExcludeNodes: []string{node2, node1},
		},
		{
			name:     "update drops lists removed from the configuration",
			existing: existing,
			config:   map[string]interface{}{},
		},
		{
			name:     "update replaces lists instead of appending to them",
			existing: existing,
			config: map[string]interface{}{
				constants.FieldKubeOVNProviderNetCustomInterfaces: []interface{}{customInterface(pnInterface, node1, node2)},
				constants.FieldKubeOVNProviderNetExcludeNodes:     []interface{}{node1, node2},
			},
			expectedCustomInterfaces: []kubeovnv1.CustomInterface{{Interface: pnInterface, Nodes: []string{node1, node2}}},
			expectedExcludeNodes:     []string{node1, node2},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			constructor := Creator(pnName)
			if tc.existing != nil {
				constructor = Updater(tc.existing.DeepCopy())
			}
			d := schema.TestResourceDataRaw(t, Schema(), tc.config)
			result, err := util.ResourceConstruct(context.Background(), d, constructor)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			pn := result.(*kubeovnv1.ProviderNetwork)

			if !slices.EqualFunc(pn.Spec.CustomInterfaces, tc.expectedCustomInterfaces, func(a, b kubeovnv1.CustomInterface) bool {
				return a.Interface == b.Interface && slices.Equal(a.Nodes, b.Nodes)
			}) {
				t.Errorf("CustomInterfaces: expected %v, got %v", tc.expectedCustomInterfaces, pn.Spec.CustomInterfaces)
			}
			if !slices.Equal(pn.Spec.ExcludeNodes, tc.expectedExcludeNodes) {
				t.Errorf("ExcludeNodes: expected %v, got %v", tc.expectedExcludeNodes, pn.Spec.ExcludeNodes)
			}
		})
	}
}

// TestProviderNetworkConstructorExchangeLinkName checks that exchange_link_name is sent as configured, false included, so
// that an update can switch it off.
func TestProviderNetworkConstructorExchangeLinkName(t *testing.T) {
	testcases := []struct {
		name     string
		existing *kubeovnv1.ProviderNetwork // nil creates the object
		value    bool
	}{
		{name: "create with true", value: true},
		{name: "create with false", value: false},
		{name: "update switches it off", existing: &kubeovnv1.ProviderNetwork{ObjectMeta: metav1.ObjectMeta{Name: pnName}, Spec: kubeovnv1.ProviderNetworkSpec{ExchangeLinkName: true, DefaultInterface: pnInterface}}, value: false},
		{name: "update switches it on", existing: &kubeovnv1.ProviderNetwork{ObjectMeta: metav1.ObjectMeta{Name: pnName}, Spec: kubeovnv1.ProviderNetworkSpec{DefaultInterface: pnInterface}}, value: true},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, Schema(), map[string]interface{}{
				constants.FieldCommonName:                         pnName,
				constants.FieldKubeOVNProviderNetDefaultInterface: pnInterface,
				constants.FieldKubeOVNProviderNetExchangeLinkName: tc.value,
			})
			constructor := Creator(pnName)
			if tc.existing != nil {
				constructor = Updater(tc.existing)
			}
			obj, err := util.ResourceConstruct(context.Background(), d, constructor)
			if err != nil {
				t.Fatal(err)
			}
			if got := obj.(*kubeovnv1.ProviderNetwork).Spec.ExchangeLinkName; got != tc.value {
				t.Errorf("exchange_link_name = %v, want %v", got, tc.value)
			}
		})
	}
}
