package importer

import (
	"testing"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/harvester/pkg/builder"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func TestVirtualMachineTemplateStateGetter(t *testing.T) {
	template := &harvsterv1.VirtualMachineTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-template",
			Namespace: "default",
			Labels: map[string]string{
				builder.LabelPrefixHarvesterTag + "env": "test",
			},
			Annotations: map[string]string{
				builder.AnnotationKeyDescription: "Test template description",
			},
		},
		Spec: harvsterv1.VirtualMachineTemplateSpec{
			DefaultVersionID: "default/test-template-v1",
			Description:      "Test template description",
		},
		Status: harvsterv1.VirtualMachineTemplateStatus{
			DefaultVersion: 1,
			LatestVersion:  3,
		},
	}

	stateGetter, err := ResourceVirtualMachineTemplateStateGetter(template)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stateGetter.ID != "default/test-template" {
		t.Errorf("ID = %q, want %q", stateGetter.ID, "default/test-template")
	}
	if stateGetter.Name != "test-template" {
		t.Errorf("Name = %q, want %q", stateGetter.Name, "test-template")
	}
	if stateGetter.ResourceType != constants.ResourceTypeVirtualMachineTemplate {
		t.Errorf("ResourceType = %q, want %q", stateGetter.ResourceType, constants.ResourceTypeVirtualMachineTemplate)
	}
	if got := stateGetter.States[constants.FieldVirtualMachineTemplateDefaultVersionID]; got != "default/test-template-v1" {
		t.Errorf("DefaultVersionID = %q, want %q", got, "default/test-template-v1")
	}
	if got := stateGetter.States[constants.FieldVirtualMachineTemplateDefaultVersion]; got != 1 {
		t.Errorf("DefaultVersion = %v, want %v", got, 1)
	}
	if got := stateGetter.States[constants.FieldVirtualMachineTemplateLatestVersion]; got != 3 {
		t.Errorf("LatestVersion = %v, want %v", got, 3)
	}
	tags := stateGetter.States[constants.FieldCommonTags].(map[string]string)
	if tags["env"] != "test" {
		t.Errorf("Tags[env] = %q, want %q", tags["env"], "test")
	}
}

func TestVirtualMachineTemplateStateGetterMinimal(t *testing.T) {
	template := &harvsterv1.VirtualMachineTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "minimal-template",
			Namespace:   "default",
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Spec:   harvsterv1.VirtualMachineTemplateSpec{},
		Status: harvsterv1.VirtualMachineTemplateStatus{},
	}

	stateGetter, err := ResourceVirtualMachineTemplateStateGetter(template)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stateGetter.ID != "default/minimal-template" {
		t.Errorf("ID = %q, want %q", stateGetter.ID, "default/minimal-template")
	}
	if got := stateGetter.States[constants.FieldVirtualMachineTemplateDefaultVersionID]; got != "" {
		t.Errorf("DefaultVersionID = %q, want empty", got)
	}
	if got := stateGetter.States[constants.FieldVirtualMachineTemplateDefaultVersion]; got != 0 {
		t.Errorf("DefaultVersion = %v, want 0", got)
	}
}
