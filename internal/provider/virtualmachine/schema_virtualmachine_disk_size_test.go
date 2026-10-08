package virtualmachine

import (
	"context"
	"testing"

	"github.com/harvester/harvester/pkg/builder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// TestDiskSizePlan checks the plan of a disk whose size is not configured:
// the provider creates it with builder.DefaultDiskSize and reads that size
// back, which must not show up as a change.
func TestDiskSizePlan(t *testing.T) {
	disk := func(size string) map[string]any {
		d := map[string]any{
			constants.FieldDiskName:    "rootdisk",
			constants.FieldVolumeImage: "default/image",
		}
		if size != "" {
			d[constants.FieldDiskSize] = size
		}
		return d
	}
	config := func(size string) map[string]any {
		return map[string]any{
			constants.FieldCommonName:         "vm",
			constants.FieldCommonNamespace:    "default",
			constants.FieldVirtualMachineDisk: []any{disk(size)},
		}
	}

	testcases := []struct {
		name       string
		configured string
		wantChange bool
	}{
		{
			name:       "size not configured",
			configured: "",
			wantChange: false,
		},
		{
			name:       "same size configured",
			configured: builder.DefaultDiskSize,
			wantChange: false,
		},
		{
			name:       "larger size configured",
			configured: "20Gi",
			wantChange: true,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// State as read back from a VM created without a size.
			stored := schema.TestResourceDataRaw(t, Schema(), config(builder.DefaultDiskSize))
			stored.SetId("default/vm")
			state := stored.State()

			diff, err := ResourceVirtualMachine().Diff(context.Background(), state, terraform.NewResourceConfigRaw(config(tc.configured)), nil)
			if err != nil {
				t.Fatal(err)
			}
			attribute := constants.FieldVirtualMachineDisk + ".0." + constants.FieldDiskSize
			changed := false
			if diff != nil {
				_, changed = diff.Attributes[attribute]
			}
			if changed != tc.wantChange {
				t.Errorf("%s in the plan = %v, want %v", attribute, changed, tc.wantChange)
			}
		})
	}
}
