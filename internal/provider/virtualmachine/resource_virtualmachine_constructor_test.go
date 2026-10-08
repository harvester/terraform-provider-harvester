package virtualmachine

import (
	"context"
	"testing"

	"github.com/harvester/harvester/pkg/builder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func TestDiskEject(t *testing.T) {
	const image = "example/image:latest"
	disk := func(name, diskType string, bootOrder int, eject bool) map[string]any {
		return map[string]any{
			constants.FieldDiskName:               name,
			constants.FieldDiskType:               diskType,
			constants.FieldDiskBus:                string(kubevirtv1.DiskBusSATA),
			constants.FieldDiskBootOrder:          bootOrder,
			constants.FieldDiskContainerImageName: image,
			constants.FieldDiskEject:              eject,
		}
	}

	tests := []struct {
		name      string
		disks     []any
		wantTrays map[string]kubevirtv1.TrayState
		wantErr   bool
	}{
		{
			name:      "ejected cd-rom has an open tray",
			disks:     []any{disk("iso", builder.DiskTypeCDRom, 1, true)},
			wantTrays: map[string]kubevirtv1.TrayState{"iso": kubevirtv1.TrayStateOpen},
		},
		{
			name:      "cd-rom without eject keeps the default tray",
			disks:     []any{disk("iso", builder.DiskTypeCDRom, 1, false)},
			wantTrays: map[string]kubevirtv1.TrayState{"iso": ""},
		},
		{
			name: "only the ejected cd-rom is opened",
			disks: []any{
				disk("rootdisk", builder.DiskTypeDisk, 1, false),
				disk("iso-1", builder.DiskTypeCDRom, 2, false),
				disk("iso-2", builder.DiskTypeCDRom, 3, true),
			},
			wantTrays: map[string]kubevirtv1.TrayState{"iso-1": "", "iso-2": kubevirtv1.TrayStateOpen},
		},
		{
			name:    "eject on a disk is refused",
			disks:   []any{disk("rootdisk", builder.DiskTypeDisk, 1, true)},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			d := schema.TestResourceDataRaw(t, Schema(), map[string]any{
				constants.FieldCommonName:         "test-vm",
				constants.FieldCommonNamespace:    "default",
				constants.FieldVirtualMachineDisk: tc.disks,
			})
			obj, err := util.ResourceConstruct(ctx, d, Creator(nil, ctx, "default", "test-vm"))
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ResourceConstruct() error: %v", err)
			}
			for _, got := range obj.(*kubevirtv1.VirtualMachine).Spec.Template.Spec.Domain.Devices.Disks {
				want, isCDRom := tc.wantTrays[got.Name]
				if !isCDRom {
					continue
				}
				if got.CDRom == nil || got.CDRom.Tray != want {
					t.Errorf("disk %s: cd-rom = %+v, want tray %q", got.Name, got.CDRom, want)
				}
			}
		})
	}
}
