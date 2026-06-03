package util

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func dummyDiffSuppress(_, _, _ string, _ *schema.ResourceData) bool { return false }
func dummyValidateDiag(_ interface{}, _ cty.Path) diag.Diagnostics  { return nil }
func dummyValidate(_ interface{}, _ string) ([]string, []error)     { return nil, nil }
func dummyStateFunc(_ interface{}) string                           { return "" }

// TestDataSourceSchemaWrap verifies that DataSourceSchemaWrap makes fields
// computed-only and clears every attribute the terraform-plugin-sdk rejects on a
// computed-only field, recursively through nested Elem schemas (list and set),
// while leaving the name/namespace query inputs configurable. The final
// InternalValidate assertion reproduces the provider startup check that
// previously failed.
func TestDataSourceSchemaWrap(t *testing.T) {
	src := map[string]*schema.Schema{
		// name/namespace are the data-source query inputs and must stay configurable.
		constants.FieldCommonName: {
			Type:     schema.TypeString,
			Required: true,
		},
		// a field exercising every attribute that is illegal on computed-only fields.
		"flag": {
			Type:                  schema.TypeString,
			Optional:              true,
			Default:               "x",
			InputDefault:          "x",
			StateFunc:             dummyStateFunc,
			DiffSuppressFunc:      dummyDiffSuppress,
			DiffSuppressOnRefresh: true,
			ValidateDiagFunc:      dummyValidateDiag,
			ConflictsWith:         []string{"flag2"},
		},
		"flag2": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: dummyValidate,
		},
		// nested set-of-resource (depth 2) to exercise the recursion.
		"nested": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"inner_set": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"deep": {
									Type:             schema.TypeInt,
									Optional:         true,
									Default:          1,
									DiffSuppressFunc: dummyDiffSuppress,
									StateFunc:        dummyStateFunc,
								},
							},
						},
					},
				},
			},
		},
	}

	wrapped := DataSourceSchemaWrap(src)

	// name/namespace stay configurable inputs (not forced to computed-only).
	if name := wrapped[constants.FieldCommonName]; !name.Required || name.Computed {
		t.Errorf("%s should remain a required input, got computed=%v required=%v", constants.FieldCommonName, name.Computed, name.Required)
	}

	computedOnly := func(t *testing.T, f *schema.Schema, field string) {
		t.Helper()
		if !f.Computed || f.Optional || f.Required {
			t.Errorf("%s: want computed-only, got computed=%v optional=%v required=%v", field, f.Computed, f.Optional, f.Required)
		}
		if f.Default != nil || f.DefaultFunc != nil || f.InputDefault != "" || f.StateFunc != nil ||
			f.ValidateFunc != nil || f.ValidateDiagFunc != nil ||
			f.DiffSuppressFunc != nil || f.DiffSuppressOnRefresh ||
			f.ConflictsWith != nil || f.AtLeastOneOf != nil || f.ExactlyOneOf != nil || f.RequiredWith != nil ||
			f.MaxItems != 0 || f.MinItems != 0 {
			t.Errorf("%s: an attribute illegal on a computed-only field was not cleared: %+v", field, f)
		}
	}
	computedOnly(t, wrapped["flag"], "flag")
	computedOnly(t, wrapped["flag2"], "flag2")
	computedOnly(t, wrapped["nested"], "nested")
	innerSet := wrapped["nested"].Elem.(*schema.Resource).Schema["inner_set"]
	computedOnly(t, innerSet, "nested.inner_set")
	computedOnly(t, innerSet.Elem.(*schema.Resource).Schema["deep"], "nested.inner_set.deep")

	// Definitive check: a data source built from the wrapped schema must pass the
	// provider's InternalValidate (the check that previously failed at startup).
	ds := &schema.Resource{Schema: wrapped}
	if err := ds.InternalValidate(nil, false); err != nil {
		t.Fatalf("InternalValidate failed for wrapped data-source schema: %v", err)
	}
}
