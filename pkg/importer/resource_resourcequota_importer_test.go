package importer

import (
	"testing"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func TestResourceResourceQuotaStateGetter(t *testing.T) {
	obj := &harvsterv1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-quota",
			Namespace:   "default",
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Spec: harvsterv1.ResourceQuotaSpec{
			SnapshotLimit: harvsterv1.SnapshotLimit{
				NamespaceTotalSnapshotSizeQuota: 107374182400,
				VMTotalSnapshotSizeQuota: map[string]int64{
					"my-vm": 53687091200,
				},
			},
		},
		Status: harvsterv1.ResourceQuotaStatus{
			SnapshotLimitStatus: harvsterv1.SnapshotLimitStatus{
				NamespaceTotalSnapshotSizeUsage: 1073741824,
				VMTotalSnapshotSizeUsage: map[string]int64{
					"my-vm": 536870912,
				},
			},
		},
	}

	stateGetter, err := ResourceResourceQuotaStateGetter(obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stateGetter.ID != "default/test-quota" {
		t.Errorf("expected ID 'default/test-quota', got '%s'", stateGetter.ID)
	}

	if stateGetter.Name != "test-quota" {
		t.Errorf("expected Name 'test-quota', got '%s'", stateGetter.Name)
	}

	if stateGetter.ResourceType != constants.ResourceTypeResourceQuota {
		t.Errorf("expected ResourceType '%s', got '%s'", constants.ResourceTypeResourceQuota, stateGetter.ResourceType)
	}

	nsQuota := stateGetter.States[constants.FieldResourceQuotaNamespaceTotalSnapshotSizeQuota].(int)
	if nsQuota != 107374182400 {
		t.Errorf("expected namespace quota 107374182400, got %d", nsQuota)
	}

	vmQuotas := stateGetter.States[constants.FieldResourceQuotaVMTotalSnapshotSizeQuota].(map[string]interface{})
	if vmQuota, ok := vmQuotas["my-vm"]; !ok || vmQuota.(int) != 53687091200 {
		t.Errorf("expected vm quota 53687091200 for 'my-vm', got %v", vmQuotas)
	}

	nsUsage := stateGetter.States[constants.FieldResourceQuotaNamespaceTotalSnapshotSizeUsage].(int)
	if nsUsage != 1073741824 {
		t.Errorf("expected namespace usage 1073741824, got %d", nsUsage)
	}

	vmUsage := stateGetter.States[constants.FieldResourceQuotaVMTotalSnapshotSizeUsage].(map[string]interface{})
	if usage, ok := vmUsage["my-vm"]; !ok || usage.(int) != 536870912 {
		t.Errorf("expected vm usage 536870912 for 'my-vm', got %v", vmUsage)
	}
}

func TestResourceResourceQuotaStateGetterEmpty(t *testing.T) {
	obj := &harvsterv1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "empty-quota",
			Namespace:   "test-ns",
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Spec:   harvsterv1.ResourceQuotaSpec{},
		Status: harvsterv1.ResourceQuotaStatus{},
	}

	stateGetter, err := ResourceResourceQuotaStateGetter(obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stateGetter.ID != "test-ns/empty-quota" {
		t.Errorf("expected ID 'test-ns/empty-quota', got '%s'", stateGetter.ID)
	}

	nsQuota := stateGetter.States[constants.FieldResourceQuotaNamespaceTotalSnapshotSizeQuota].(int)
	if nsQuota != 0 {
		t.Errorf("expected namespace quota 0, got %d", nsQuota)
	}

	vmQuotas := stateGetter.States[constants.FieldResourceQuotaVMTotalSnapshotSizeQuota].(map[string]interface{})
	if len(vmQuotas) != 0 {
		t.Errorf("expected empty vm quotas, got %v", vmQuotas)
	}
}
