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
		constants.FieldVirtualMachineInstanceNodeName: {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
	util.NamespacedSchemaWrap(s, false)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
