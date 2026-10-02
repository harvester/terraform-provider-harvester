package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceKubeOVNVlanStateGetter(t *testing.T) {
	testcases := []struct {
		name           string
		vlan           *kubeovnv1.Vlan
		expectedID     string
		expectedVlanID int
		expectedProv   string
		expectedSubs   int
	}{
		{
			name: "vlan with provider and subnets",
			vlan: &kubeovnv1.Vlan{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-vlan",
				},
				Spec: kubeovnv1.VlanSpec{
					ID:       100,
					Provider: "provider-net1",
				},
				Status: kubeovnv1.VlanStatus{
					Subnets: []string{"subnet1", "subnet2"},
				},
			},
			expectedID:     helper.BuildID("", "test-vlan"),
			expectedVlanID: 100,
			expectedProv:   "provider-net1",
			expectedSubs:   2,
		},
		{
			name: "empty vlan",
			vlan: &kubeovnv1.Vlan{
				ObjectMeta: metav1.ObjectMeta{
					Name: "empty-vlan",
				},
				Spec: kubeovnv1.VlanSpec{},
			},
			expectedID:     helper.BuildID("", "empty-vlan"),
			expectedVlanID: 0,
			expectedProv:   "",
			expectedSubs:   0,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNVlanStateGetter(tc.vlan)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNVlan {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeKubeOVNVlan, getter.ResourceType)
			}

			vlanID := getter.States[constants.FieldKubeOVNVlanID].(int)
			if vlanID != tc.expectedVlanID {
				t.Errorf("VlanID: expected %d, got %d", tc.expectedVlanID, vlanID)
			}

			prov := getter.States[constants.FieldKubeOVNVlanProvider].(string)
			if prov != tc.expectedProv {
				t.Errorf("Provider: expected %q, got %q", tc.expectedProv, prov)
			}

			var subsLen int
			if subs := getter.States[constants.FieldKubeOVNVlanStatusSubnets]; subs != nil {
				subsLen = len(subs.([]string))
			}
			if subsLen != tc.expectedSubs {
				t.Errorf("StatusSubnets: expected %d, got %d", tc.expectedSubs, subsLen)
			}
		})
	}
}
