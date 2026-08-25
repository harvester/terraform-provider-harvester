package addon

import (
	"context"
	"testing"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// TestConstructorDisableAddon verifies that `enabled = false` is applied on
// update. The enabled processor must be required: non-required processors are
// read with GetOk which skips zero values, so a false value would be ignored
// and the addon could never be disabled in place.
func TestConstructorDisableAddon(t *testing.T) {
	d := schema.TestResourceDataRaw(t, Schema(), map[string]interface{}{
		constants.FieldCommonName:   "vm-import-controller",
		constants.FieldAddonEnabled: false,
	})

	oldAddon := &harvsterv1.Addon{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vm-import-controller",
			Namespace: "harvester-system",
		},
		Spec: harvsterv1.AddonSpec{
			Enabled: true,
		},
	}

	result, err := util.ResourceConstruct(context.Background(), d, Updater(oldAddon))
	if err != nil {
		t.Fatalf("ResourceConstruct() error: %v", err)
	}
	if updated := result.(*harvsterv1.Addon); updated.Spec.Enabled {
		t.Errorf("Spec.Enabled = true after constructing with enabled=false, want false")
	}
}
