package kubeovn_vlan

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

const (
	vlanName     = "test-vlan"
	vlanProvider = "test-pn"
	vlanID       = 100
	otherVlanID  = 200
)

func TestVlanConstructor(t *testing.T) {
	existing := &kubeovnv1.Vlan{
		ObjectMeta: metav1.ObjectMeta{Name: vlanName},
		Spec:       kubeovnv1.VlanSpec{ID: vlanID, Provider: vlanProvider},
		Status:     kubeovnv1.VlanStatus{Subnets: []string{"subnet1"}},
	}

	testcases := []struct {
		name     string
		existing *kubeovnv1.Vlan // nil creates the VLAN
		id       int
	}{
		{name: "create tagged", id: vlanID},
		{name: "create untagged", id: 0},
		{name: "update changes the tag", existing: existing, id: otherVlanID},
		{name: "update to untagged", existing: existing, id: 0},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			constructor := Creator(vlanName)
			if tc.existing != nil {
				constructor = Updater(tc.existing.DeepCopy())
			}
			d := schema.TestResourceDataRaw(t, Schema(), map[string]interface{}{
				constants.FieldCommonName:          vlanName,
				constants.FieldKubeOVNVlanID:       tc.id,
				constants.FieldKubeOVNVlanProvider: vlanProvider,
			})
			result, err := util.ResourceConstruct(context.Background(), d, constructor)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			vlan := result.(*kubeovnv1.Vlan)
			if vlan.Name != vlanName {
				t.Errorf("name = %q, want %q", vlan.Name, vlanName)
			}
			if vlan.Spec.ID != tc.id {
				t.Errorf("spec.id = %d, want %d", vlan.Spec.ID, tc.id)
			}
			if vlan.Spec.Provider != vlanProvider {
				t.Errorf("spec.provider = %q, want %q", vlan.Spec.Provider, vlanProvider)
			}
			if tc.existing != nil && len(vlan.Status.Subnets) != len(tc.existing.Status.Subnets) {
				t.Errorf("status.subnets = %v, want %v kept", vlan.Status.Subnets, tc.existing.Status.Subnets)
			}
		})
	}
}
