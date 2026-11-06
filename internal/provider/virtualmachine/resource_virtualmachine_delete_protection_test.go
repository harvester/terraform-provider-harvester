package virtualmachine

import (
	"context"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

const (
	labelTrue  = "true"
	labelFalse = "false"
)

// TestDeleteProtectionFromRawConfig checks that an explicit false is told
// apart from an unset field.
func TestDeleteProtectionFromRawConfig(t *testing.T) {
	objectType := cty.Object(map[string]cty.Type{
		constants.FieldCommonName:                     cty.String,
		constants.FieldVirtualMachineDeleteProtection: cty.Bool,
	})
	config := func(v cty.Value) cty.Value {
		return cty.ObjectVal(map[string]cty.Value{
			constants.FieldCommonName:                     cty.StringVal("vm"),
			constants.FieldVirtualMachineDeleteProtection: v,
		})
	}
	testcases := []struct {
		name string
		raw  cty.Value
		want *bool
	}{
		{name: "no config", raw: cty.NullVal(objectType), want: nil},
		{name: "unset", raw: config(cty.NullVal(cty.Bool)), want: nil},
		{name: "unknown", raw: config(cty.UnknownVal(cty.Bool)), want: nil},
		{name: "enabled", raw: config(cty.True), want: new(true)},
		{name: "disabled", raw: config(cty.False), want: new(false)},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := deleteProtectionFromRawConfig(tc.raw)
			switch {
			case tc.want == nil && got != nil:
				t.Errorf("got %v, want unset", *got)
			case tc.want != nil && got == nil:
				t.Errorf("got unset, want %v", *tc.want)
			case tc.want != nil && *got != *tc.want:
				t.Errorf("got %v, want %v", *got, *tc.want)
			}
		})
	}

	// Without a plan, as in schema.TestResourceDataRaw, there is no raw config.
	d := schema.TestResourceDataRaw(t, ResourceVirtualMachine().Schema, map[string]interface{}{})
	if got := deleteProtectionFromConfig(d); got != nil {
		t.Errorf("no raw config: got %v, want unset", *got)
	}
}

func TestSetDeleteProtectionLabel(t *testing.T) {
	testcases := []struct {
		name    string
		labels  map[string]string
		enabled *bool
		want    string
		wantSet bool
	}{
		{name: "unset keeps a missing label missing", enabled: nil, wantSet: false},
		{name: "unset keeps the existing value", labels: map[string]string{constants.LabelVMDeleteProtection: labelTrue}, enabled: nil, want: labelTrue, wantSet: true},
		{name: "enable", enabled: new(true), want: labelTrue, wantSet: true},
		{name: "disable overrides true", labels: map[string]string{constants.LabelVMDeleteProtection: "True"}, enabled: new(false), want: labelFalse, wantSet: true},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			vm := &kubevirtv1.VirtualMachine{ObjectMeta: metav1.ObjectMeta{Labels: tc.labels}}
			setDeleteProtectionLabel(vm, tc.enabled)
			got, ok := vm.Labels[constants.LabelVMDeleteProtection]
			if ok != tc.wantSet || got != tc.want {
				t.Errorf("label = %q (set %v), want %q (set %v)", got, ok, tc.want, tc.wantSet)
			}
		})
	}
}

// TestUpdaterKeepsDeleteProtectionLabel checks that the label clean-up done
// for `labels` does not remove the delete protection label on update.
func TestUpdaterKeepsDeleteProtectionLabel(t *testing.T) {
	vm := &kubevirtv1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
			Labels: map[string]string{
				constants.LabelVMDeleteProtection: labelTrue,
				"app":                             "removed-from-config",
			},
		},
		Spec: kubevirtv1.VirtualMachineSpec{Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{}},
	}
	Updater(nil, context.Background(), vm).Setup()
	if got := vm.Labels[constants.LabelVMDeleteProtection]; got != labelTrue {
		t.Errorf("delete protection label = %q, want \"true\"", got)
	}
	if _, ok := vm.Labels["app"]; ok {
		t.Errorf("user label still present, want it removed before the processors run")
	}
}

func TestDeleteProtectionError(t *testing.T) {
	protected := &kubevirtv1.VirtualMachine{ObjectMeta: metav1.ObjectMeta{
		Namespace: "vm-ns", Name: "vm",
		Labels: map[string]string{constants.LabelVMDeleteProtection: "True"},
	}}
	if err := deleteProtectionError(protected); err == nil {
		t.Error("protected VM: got no error")
	}
	for _, value := range []string{labelFalse, "yes", ""} {
		vm := &kubevirtv1.VirtualMachine{ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{constants.LabelVMDeleteProtection: value},
		}}
		if err := deleteProtectionError(vm); err != nil {
			t.Errorf("label %q: got %v, want no error", value, err)
		}
	}
}
