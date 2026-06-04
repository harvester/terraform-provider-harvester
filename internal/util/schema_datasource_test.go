package util

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestDataSourceSchemaWrapClearsValidationFuncs ensures data-source fields become
// computed-only and that DiffSuppressFunc/ValidateDiagFunc (illegal on
// computed-only fields, would fail provider.InternalValidate) are cleared.
func TestDataSourceSchemaWrapClearsValidationFuncs(t *testing.T) {
	s := map[string]*schema.Schema{
		"foo": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "bar",
			DiffSuppressFunc: func(_, _, _ string, _ *schema.ResourceData) bool { return false },
			ValidateDiagFunc: func(_ interface{}, _ cty.Path) diag.Diagnostics { return nil },
			ValidateFunc:     func(_ interface{}, _ string) ([]string, []error) { return nil, nil },
			ConflictsWith:    []string{"baz"},
		},
	}
	DataSourceSchemaWrap(s)
	f := s["foo"]
	if !f.Computed || f.Optional || f.Required {
		t.Errorf("field should be computed-only, got computed=%v optional=%v required=%v", f.Computed, f.Optional, f.Required)
	}
	if f.DiffSuppressFunc != nil {
		t.Error("DiffSuppressFunc should be cleared")
	}
	if f.ValidateDiagFunc != nil {
		t.Error("ValidateDiagFunc should be cleared")
	}
	if f.ValidateFunc != nil || f.ConflictsWith != nil || f.Default != nil {
		t.Error("ValidateFunc/ConflictsWith/Default should be cleared")
	}
}
