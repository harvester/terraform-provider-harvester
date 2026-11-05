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

func TestResourceKubeOVNVlanStateGetter(t *testing.T) {
	const (
		vlanName = "test-vlan"
		provider = "test-pn"
		subnetA  = "subnet-a"
		subnetB  = "subnet-b"
	)

	testcases := []struct {
		name     string
		vlan     *kubeovnv1.Vlan
		expected map[string]interface{}
	}{
		{
			name: "tagged VLAN with subnets",
			vlan: &kubeovnv1.Vlan{
				ObjectMeta: metav1.ObjectMeta{
					Name:        vlanName,
					Labels:      map[string]string{builder.LabelPrefixHarvesterTag + testKubeOVNTagKey: testKubeOVNTagValue, testKubeOVNLabelKey: testKubeOVNLabelValue},
					Annotations: map[string]string{builder.AnnotationKeyDescription: "storage VLAN"},
				},
				Spec:   kubeovnv1.VlanSpec{ID: 100, Provider: provider},
				Status: kubeovnv1.VlanStatus{Subnets: []string{subnetB, subnetA}},
			},
			expected: map[string]interface{}{
				constants.FieldCommonState:               constants.StateCommonActive,
				constants.FieldCommonName:                vlanName,
				constants.FieldCommonDescription:         "storage VLAN",
				constants.FieldCommonTags:                map[string]string{testKubeOVNTagKey: testKubeOVNTagValue},
				constants.FieldCommonLabels:              map[string]string{testKubeOVNLabelKey: testKubeOVNLabelValue},
				constants.FieldKubeOVNVlanID:             100,
				constants.FieldKubeOVNVlanProvider:       provider,
				constants.FieldKubeOVNVlanStatusSubnets:  []string{subnetB, subnetA},
				constants.FieldKubeOVNVlanStatusConflict: false,
			},
		},
		{
			name: "untagged VLAN in conflict",
			vlan: &kubeovnv1.Vlan{
				ObjectMeta: metav1.ObjectMeta{Name: vlanName},
				Spec:       kubeovnv1.VlanSpec{Provider: provider},
				Status:     kubeovnv1.VlanStatus{Conflict: true},
			},
			expected: map[string]interface{}{
				constants.FieldCommonState:               constants.StateCommonActive,
				constants.FieldCommonName:                vlanName,
				constants.FieldCommonDescription:         "",
				constants.FieldCommonTags:                map[string]string{},
				constants.FieldCommonLabels:              map[string]string{},
				constants.FieldKubeOVNVlanID:             0,
				constants.FieldKubeOVNVlanProvider:       provider,
				constants.FieldKubeOVNVlanStatusSubnets:  []string(nil),
				constants.FieldKubeOVNVlanStatusConflict: true,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNVlanStateGetter(tc.vlan)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if getter.ID != helper.BuildID("", vlanName) || getter.Name != vlanName {
				t.Errorf("ID, Name = %q, %q, want %q, %q", getter.ID, getter.Name, helper.BuildID("", vlanName), vlanName)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNVlan {
				t.Errorf("ResourceType = %q, want %q", getter.ResourceType, constants.ResourceTypeKubeOVNVlan)
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
