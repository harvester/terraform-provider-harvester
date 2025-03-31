package virtualmachine

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldVirtualMachineMachineType: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		constants.FieldVirtualMachineHostname: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		constants.FieldVirtualMachineReservedMemory: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.FieldVirtualMachineRestartAfterUpdate: {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "restart vm after the vm is updated",
		},
		constants.FieldVirtualMachineRunStrategy: {
			Type:     schema.TypeString,
			Optional: true,
			Default:  string(kubevirtv1.RunStrategyRerunOnFailure),
			ValidateFunc: validation.StringInSlice([]string{
				string(kubevirtv1.RunStrategyAlways),
				string(kubevirtv1.RunStrategyManual),
				string(kubevirtv1.RunStrategyHalted),
				string(kubevirtv1.RunStrategyRerunOnFailure),
			}, false),
			Description: "more info: https://kubevirt.io/user-guide/virtual_machines/run_strategies/",
		},
		constants.FieldVirtualMachineStart: {
			Type:     schema.TypeBool,
			Optional: true,
			Deprecated: fmt.Sprintf(`
please use %s instead of this deprecated field:
	%s = true  ==>  %s = "%s"
	%s = false  ==>  %s = "%s"
`, constants.FieldVirtualMachineRunStrategy,
				constants.FieldVirtualMachineStart, constants.FieldVirtualMachineRunStrategy, kubevirtv1.RunStrategyRerunOnFailure,
				constants.FieldVirtualMachineStart, constants.FieldVirtualMachineRunStrategy, kubevirtv1.RunStrategyHalted),
		},
		constants.FieldVirtualMachineCPU: {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  1,
		},
		constants.FieldVirtualMachineMemory: {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "1Gi",
		},
		constants.FieldVirtualMachineSSHKeys: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "The `ssh_keys` are added to `cloudinit.user_data` if:\n" +
				"1. Both `cloudinit.user_data_base64` and `cloudinit.user_data_secret_name` are empty.\n" +
				"2. There is no `ssh_authorized_keys` field in `cloudinit.user_data`.",
		},
		constants.FieldVirtualMachineCloudInit: {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: resourceCloudInitSchema(),
			},
		},
		constants.FieldVirtualMachineDisk: {
			Type:     schema.TypeList,
			Required: true,
			MinItems: 1,
			Elem: &schema.Resource{
				Schema: resourceDiskSchema(),
			},
		},
		constants.FieldVirtualMachineNetworkInterface: {
			Type:     schema.TypeList,
			Required: true,
			MinItems: 1,
			Elem: &schema.Resource{
				Schema: resourceNetworkInterfaceSchema(),
			},
		},
		constants.FieldVirtualMachineInput: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: resourceInputSchema(),
			},
		},
		constants.FieldVirtualMachineTPM: {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: resourceTPMSchema(),
			},
		},
		constants.FieldVirtualMachineInstanceNodeName: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldVirtualMachineEFI: {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		constants.FieldVirtualMachineSecureBoot: {
			Type:        schema.TypeBool,
			Description: "EFI must be enabled to use this feature",
			Optional:    true,
			Default:     false,
		},
		constants.FieldVirtualMachineCPUPinning: {
			Type:        schema.TypeBool,
			Description: "To enable VM CPU pinning, ensure that at least one node has the CPU manager enabled",
			Optional:    true,
			Default:     false,
		},
		constants.FieldVirtualMachineIsolateEmulatorThread: {
			Type:        schema.TypeBool,
			Description: "To enable isolate emulator thread, ensure that at least one node has the CPU manager enabled, also VM CPU pinning must be enabled. Note that enable option will allocate an additional dedicated CPU.",
			Optional:    true,
			Default:     false,
		},
	}
	util.NamespacedSchemaWrap(s, false)
	s[constants.FieldCommonTags].Description = "The `ssh-user` is added to `cloudinit.user_data` if:\n" +
		"1. Both `cloudinit.user_data_base64` and `cloudinit.user_data_secret_name` are empty.\n" +
		"2. There is no `user` field in `cloudinit.user_data`.\n"
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
