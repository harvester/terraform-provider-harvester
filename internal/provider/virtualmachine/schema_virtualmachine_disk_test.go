package virtualmachine

import (
	"context"
	"testing"

	"github.com/harvester/harvester/pkg/builder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// TestEmptyCDRomPlan checks the plan of a disk whose volume fields are read
// back empty. An empty cd-rom has no volume, so the importer has nothing to
// read them from: the configured values must not show up as a change. A disk
// backed by a volume must still report them.
func TestEmptyCDRomPlan(t *testing.T) {
	const image = "default/image"
	volumeFields := map[string]any{
		constants.FieldDiskSize:               "1Gi",
		constants.FieldDiskAutoDelete:         true,
		constants.FieldDiskHotPlug:            true,
		constants.FieldVolumeStorageClassName: "longhorn",
		constants.FieldVolumeMode:             builder.PersistentVolumeModeBlock,
		constants.FieldVolumeAccessMode:       builder.PersistentVolumeAccessModeReadWriteMany,
	}
	disk := func(diskType, image string, configured bool) map[string]any {
		d := map[string]any{
			constants.FieldDiskName:    "cd1",
			constants.FieldDiskType:    diskType,
			constants.FieldDiskBus:     builder.DiskBusSata,
			constants.FieldVolumeImage: image,
		}
		if configured {
			for field, value := range volumeFields {
				d[field] = value
			}
		}
		return d
	}
	config := func(d map[string]any) map[string]any {
		return map[string]any{
			constants.FieldCommonName:         "vm",
			constants.FieldCommonNamespace:    "default",
			constants.FieldVirtualMachineDisk: []any{d},
		}
	}

	testcases := []struct {
		name       string
		stored     map[string]any
		configured map[string]any
		wantChange bool
	}{
		{
			name:       "empty cd-rom",
			stored:     disk(builder.DiskTypeCDRom, "", false),
			configured: disk(builder.DiskTypeCDRom, "", true),
			wantChange: false,
		},
		{
			name:       "image inserted into an empty cd-rom",
			stored:     disk(builder.DiskTypeCDRom, "", false),
			configured: disk(builder.DiskTypeCDRom, image, true),
			wantChange: true,
		},
		{
			name:       "cd-rom with an image",
			stored:     disk(builder.DiskTypeCDRom, image, false),
			configured: disk(builder.DiskTypeCDRom, image, true),
			wantChange: true,
		},
		{
			name:       "disk without image",
			stored:     disk(builder.DiskTypeDisk, "", false),
			configured: disk(builder.DiskTypeDisk, "", true),
			wantChange: true,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			stored := schema.TestResourceDataRaw(t, Schema(), config(tc.stored))
			stored.SetId("default/vm")

			diff, err := ResourceVirtualMachine().Diff(context.Background(), stored.State(), terraform.NewResourceConfigRaw(config(tc.configured)), nil)
			if err != nil {
				t.Fatal(err)
			}
			for field := range volumeFields {
				attribute := constants.FieldVirtualMachineDisk + ".0." + field
				changed := false
				if diff != nil {
					_, changed = diff.Attributes[attribute]
				}
				if changed != tc.wantChange {
					t.Errorf("%s in the plan = %v, want %v", attribute, changed, tc.wantChange)
				}
			}
		})
	}
}
