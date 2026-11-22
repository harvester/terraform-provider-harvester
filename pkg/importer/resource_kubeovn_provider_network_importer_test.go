package importer

import (
	"reflect"
	"testing"

	"github.com/harvester/harvester/pkg/builder"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

// Tag and label shared by the provider network and VLAN state getter tests.
const (
	testKubeOVNTagKey     = "env"
	testKubeOVNTagValue   = "prod"
	testKubeOVNLabelKey   = "team"
	testKubeOVNLabelValue = "net"
)

func TestResourceKubeOVNProviderNetworkStateGetter(t *testing.T) {
	const (
		pnName = "test-pn"
		eth0   = "eth0"
		eth1   = "eth1"
		bond0  = "bond0"
		node1  = "node1"
		node2  = "node2"
		node3  = "node3"
	)

	testcases := []struct {
		name     string
		pn       *kubeovnv1.ProviderNetwork
		expected map[string]interface{}
	}{
		{
			name: "every field set, ready",
			pn: &kubeovnv1.ProviderNetwork{
				ObjectMeta: metav1.ObjectMeta{
					Name:        pnName,
					Labels:      map[string]string{builder.LabelPrefixHarvesterTag + testKubeOVNTagKey: testKubeOVNTagValue, testKubeOVNLabelKey: testKubeOVNLabelValue},
					Annotations: map[string]string{builder.AnnotationKeyDescription: "uplink"},
				},
				Spec: kubeovnv1.ProviderNetworkSpec{
					DefaultInterface: eth0,
					CustomInterfaces: []kubeovnv1.CustomInterface{
						{Interface: eth1, Nodes: []string{node2, node1}},
						{Interface: bond0, Nodes: []string{node3}},
					},
					ExcludeNodes:     []string{node3, node2},
					ExchangeLinkName: true,
				},
				Status: kubeovnv1.ProviderNetworkStatus{
					Ready:      true,
					ReadyNodes: []string{node1, node2},
					Vlans:      []string{"vlan100", "vlan200"},
				},
			},
			expected: map[string]interface{}{
				constants.FieldCommonState:                        constants.StateCommonReady,
				constants.FieldCommonName:                         pnName,
				constants.FieldCommonDescription:                  "uplink",
				constants.FieldCommonTags:                         map[string]string{testKubeOVNTagKey: testKubeOVNTagValue},
				constants.FieldCommonLabels:                       map[string]string{testKubeOVNLabelKey: testKubeOVNLabelValue},
				constants.FieldKubeOVNProviderNetDefaultInterface: eth0,
				constants.FieldKubeOVNProviderNetCustomInterfaces: []map[string]interface{}{
					{constants.FieldKubeOVNCustomInterfaceInterface: eth1, constants.FieldKubeOVNCustomInterfaceNodes: []string{node2, node1}},
					{constants.FieldKubeOVNCustomInterfaceInterface: bond0, constants.FieldKubeOVNCustomInterfaceNodes: []string{node3}},
				},
				constants.FieldKubeOVNProviderNetExcludeNodes:        []string{node3, node2},
				constants.FieldKubeOVNProviderNetExchangeLinkName:    true,
				constants.FieldKubeOVNProviderNetStatusReady:         true,
				constants.FieldKubeOVNProviderNetStatusReadyNodes:    []string{node1, node2},
				constants.FieldKubeOVNProviderNetStatusNotReadyNodes: []string(nil),
				constants.FieldKubeOVNProviderNetStatusVlans:         []string{"vlan100", "vlan200"},
			},
		},
		{
			name: "interface missing on a node, not ready",
			pn: &kubeovnv1.ProviderNetwork{
				ObjectMeta: metav1.ObjectMeta{Name: pnName},
				Spec:       kubeovnv1.ProviderNetworkSpec{DefaultInterface: eth1},
				Status: kubeovnv1.ProviderNetworkStatus{
					ReadyNodes:    []string{node1},
					NotReadyNodes: []string{node2},
				},
			},
			expected: map[string]interface{}{
				constants.FieldCommonState:                           constants.StateCommonActive,
				constants.FieldCommonName:                            pnName,
				constants.FieldCommonDescription:                     "",
				constants.FieldCommonTags:                            map[string]string{},
				constants.FieldCommonLabels:                          map[string]string{},
				constants.FieldKubeOVNProviderNetDefaultInterface:    eth1,
				constants.FieldKubeOVNProviderNetCustomInterfaces:    []map[string]interface{}{},
				constants.FieldKubeOVNProviderNetExcludeNodes:        []string(nil),
				constants.FieldKubeOVNProviderNetExchangeLinkName:    false,
				constants.FieldKubeOVNProviderNetStatusReady:         false,
				constants.FieldKubeOVNProviderNetStatusReadyNodes:    []string{node1},
				constants.FieldKubeOVNProviderNetStatusNotReadyNodes: []string{node2},
				constants.FieldKubeOVNProviderNetStatusVlans:         []string(nil),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNProviderNetworkStateGetter(tc.pn)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if getter.ID != helper.BuildID("", pnName) || getter.Name != pnName {
				t.Errorf("ID, Name = %q, %q, want %q, %q", getter.ID, getter.Name, helper.BuildID("", pnName), pnName)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNProviderNetwork {
				t.Errorf("ResourceType = %q, want %q", getter.ResourceType, constants.ResourceTypeKubeOVNProviderNetwork)
			}
			if !reflect.DeepEqual(getter.States, tc.expected) {
				for key := range mergeKeys(getter.States, tc.expected) {
					if !reflect.DeepEqual(getter.States[key], tc.expected[key]) {
						t.Errorf("%s = %#v, want %#v", key, getter.States[key], tc.expected[key])
					}
				}
			}
		})
	}
}

// mergeKeys returns the keys of both maps, so that a missing or an extra
// state is reported too.
func mergeKeys(a, b map[string]interface{}) map[string]struct{} {
	keys := make(map[string]struct{}, len(a)+len(b))
	for key := range a {
		keys[key] = struct{}{}
	}
	for key := range b {
		keys[key] = struct{}{}
	}
	return keys
}
