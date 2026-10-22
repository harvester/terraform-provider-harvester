package virtualmachine

import (
	"context"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func TestWarnRestartWithHotPlug(t *testing.T) {
	disk := func(name string, hotPlug cty.Value) cty.Value {
		return cty.ObjectVal(map[string]cty.Value{
			constants.FieldDiskName:    cty.StringVal(name),
			constants.FieldDiskHotPlug: hotPlug,
		})
	}
	rootDisk := disk("rootdisk", cty.NullVal(cty.Bool))
	cdrom := disk("cd1", cty.True)
	data := disk("data", cty.True)
	always := cty.StringVal(string(kubevirtv1.RunStrategyAlways))

	testcases := []struct {
		name            string
		restart         cty.Value
		runStrategy     cty.Value
		disks           []cty.Value
		expectedWarning bool
	}{
		{
			name:            "restart with a hot-plugged disk",
			restart:         cty.True,
			disks:           []cty.Value{rootDisk, cdrom},
			expectedWarning: true,
		},
		{
			name:            "restart with several hot-plugged disks warns once",
			restart:         cty.True,
			runStrategy:     always,
			disks:           []cty.Value{rootDisk, cdrom, data},
			expectedWarning: true,
		},
		{
			name:    "restart without hot-plugged disk",
			restart: cty.True,
			disks:   []cty.Value{rootDisk, disk("data", cty.False)},
		},
		{
			name:    "no restart",
			restart: cty.False,
			disks:   []cty.Value{rootDisk, cdrom},
		},
		{
			name:    "restart omitted",
			restart: cty.NullVal(cty.Bool),
			disks:   []cty.Value{rootDisk, cdrom},
		},
		{
			name:        "run strategy without restart",
			restart:     cty.True,
			runStrategy: cty.StringVal(string(kubevirtv1.RunStrategyHalted)),
			disks:       []cty.Value{rootDisk, cdrom},
		},
		{
			name:    "restart known after apply",
			restart: cty.UnknownVal(cty.Bool),
			disks:   []cty.Value{rootDisk, cdrom},
		},
		{
			name:    "hot_plug known after apply",
			restart: cty.True,
			disks:   []cty.Value{rootDisk, disk("data", cty.UnknownVal(cty.Bool))},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			runStrategy := tc.runStrategy
			if runStrategy == cty.NilVal {
				runStrategy = cty.NullVal(cty.String)
			}
			// The SDK passes the whole configuration, every attribute of the
			// schema included: fill the ones the test leaves out with null.
			config, err := ResourceVirtualMachine().CoreConfigSchema().CoerceValue(cty.ObjectVal(map[string]cty.Value{
				constants.FieldCommonName:                       cty.StringVal("vm"),
				constants.FieldVirtualMachineRestartAfterUpdate: tc.restart,
				constants.FieldVirtualMachineRunStrategy:        runStrategy,
				constants.FieldVirtualMachineDisk:               cty.ListVal(tc.disks),
			}))
			if err != nil {
				t.Fatal(err)
			}
			resp := &schema.ValidateResourceConfigFuncResponse{}
			warnRestartWithHotPlug(context.Background(), schema.ValidateResourceConfigFuncRequest{RawConfig: config}, resp)

			if !tc.expectedWarning {
				if len(resp.Diagnostics) != 0 {
					t.Errorf("diagnostics = %v, want none", resp.Diagnostics)
				}
				return
			}
			if len(resp.Diagnostics) != 1 || resp.Diagnostics[0].Severity != diag.Warning {
				t.Fatalf("diagnostics = %v, want one warning", resp.Diagnostics)
			}
			if !resp.Diagnostics[0].AttributePath.Equals(cty.GetAttrPath(constants.FieldVirtualMachineRestartAfterUpdate)) {
				t.Errorf("attribute path = %v, want %s", resp.Diagnostics[0].AttributePath, constants.FieldVirtualMachineRestartAfterUpdate)
			}
		})
	}
}
