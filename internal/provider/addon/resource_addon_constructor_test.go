package addon

import (
	"context"
	"strings"
	"testing"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	harvsterutil "github.com/harvester/harvester/pkg/util"
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

// TestConstructorCreateCustomAddon verifies the create path: repo, chart and
// version are mandatory, and the created object carries the auto-delete
// annotation and the experimental label required by the Harvester webhook to
// allow its deletion.
func TestConstructorCreateCustomAddon(t *testing.T) {
	t.Run("missing repo/chart/version is rejected with guidance", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, Schema(), map[string]interface{}{
			constants.FieldCommonName: "tf-test-missing",
		})
		_, err := util.ResourceConstruct(context.Background(), d, Creator("harvester-system", "tf-test-missing"))
		if err == nil || !strings.Contains(err.Error(), "requires repo, chart and version") {
			t.Fatalf("ResourceConstruct() error = %v, want the repo/chart/version guidance", err)
		}
	})

	t.Run("complete custom addon carries ownership annotation and experimental label", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, Schema(), map[string]interface{}{
			constants.FieldCommonName:   "tf-test-dummy",
			constants.FieldAddonRepo:    "http://example.invalid/charts",
			constants.FieldAddonChart:   "tf-test-dummy-chart",
			constants.FieldAddonVersion: "0.0.1",
		})
		result, err := util.ResourceConstruct(context.Background(), d, Creator("harvester-system", "tf-test-dummy"))
		if err != nil {
			t.Fatalf("ResourceConstruct() error: %v", err)
		}
		created := result.(*harvsterv1.Addon)
		if created.Spec.Repo != "http://example.invalid/charts" ||
			created.Spec.Chart != "tf-test-dummy-chart" ||
			created.Spec.Version != "0.0.1" {
			t.Errorf("spec repo/chart/version not applied: %+v", created.Spec)
		}
		if created.Annotations[constants.AnnotationAddonAutoDelete] != enabledLabelValue {
			t.Errorf("missing auto-delete annotation: %v", created.Annotations)
		}
		if created.Labels[harvsterutil.AddonExperimentalLabel] != enabledLabelValue {
			t.Errorf("missing experimental label: %v", created.Labels)
		}
	})
}

// TestConstructorChartImmutable verifies the client-side guard mirroring the
// Harvester webhook rule that the chart of an existing addon cannot change.
func TestConstructorChartImmutable(t *testing.T) {
	d := schema.TestResourceDataRaw(t, Schema(), map[string]interface{}{
		constants.FieldCommonName: "vm-import-controller",
		constants.FieldAddonChart: "another-chart",
	})
	oldAddon := &harvsterv1.Addon{
		ObjectMeta: metav1.ObjectMeta{Name: "vm-import-controller", Namespace: "harvester-system"},
		Spec:       harvsterv1.AddonSpec{Chart: "vm-import-controller"},
	}
	_, err := util.ResourceConstruct(context.Background(), d, Updater(oldAddon))
	if err == nil || !strings.Contains(err.Error(), "chart is immutable") {
		t.Fatalf("ResourceConstruct() error = %v, want chart immutability error", err)
	}
}

// TestConstructorPreservesAddonLabels verifies that addon.harvesterhci.io/*
// labels (notably the experimental label) survive an update even though the
// shared Labels processor wipes labels outside its known prefixes.
func TestConstructorPreservesAddonLabels(t *testing.T) {
	d := schema.TestResourceDataRaw(t, Schema(), map[string]interface{}{
		constants.FieldCommonName: "tf-test-dummy",
		constants.FieldCommonTags: map[string]interface{}{"team": "infra"},
	})
	oldAddon := &harvsterv1.Addon{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tf-test-dummy",
			Namespace: "harvester-system",
			Labels: map[string]string{
				harvsterutil.AddonExperimentalLabel: "true",
			},
		},
		Spec: harvsterv1.AddonSpec{Chart: "tf-test-dummy-chart"},
	}
	result, err := util.ResourceConstruct(context.Background(), d, Updater(oldAddon))
	if err != nil {
		t.Fatalf("ResourceConstruct() error: %v", err)
	}
	updated := result.(*harvsterv1.Addon)
	if updated.Labels[harvsterutil.AddonExperimentalLabel] != enabledLabelValue {
		t.Errorf("experimental label lost on update: %v", updated.Labels)
	}
	if updated.Labels["tag.harvesterhci.io/team"] != "infra" {
		t.Errorf("user tag not applied: %v", updated.Labels)
	}
}
