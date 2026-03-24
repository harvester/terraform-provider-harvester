package virtualmachinetemplateversion

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/internal/provider/virtualmachine"
	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldVirtualMachineTemplateVersionTemplateID: {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Template ID in the format namespace/name",
		},
		constants.FieldVirtualMachineTemplateVersionImageID: {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "Image ID in the format namespace/name",
		},
		constants.FieldVirtualMachineTemplateVersionKeyPairIDs: {
			Type:     schema.TypeList,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "Key pair IDs in the format namespace/name",
		},
		constants.FieldVirtualMachineTemplateVersionVersion: {
			Type:     schema.TypeInt,
			Computed: true,
		},
		constants.FieldVirtualMachineCPU: {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
			Default:  1,
		},
		constants.FieldVirtualMachineCPUModel: {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "CPU model for the virtual machine",
		},
		constants.FieldVirtualMachineMemory: {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
			Default:  "1Gi",
		},
		constants.FieldVirtualMachineRequests: {
			Type:     schema.TypeList,
			Optional: true,
			ForceNew: true,
			Computed: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldRequestsCPU: {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					constants.FieldRequestsMemory: {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
				},
			},
		},
		constants.FieldVirtualMachineMachineType: {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
			Computed: true,
		},
		constants.FieldVirtualMachineHostname: {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
			Computed: true,
		},
		constants.FieldVirtualMachineReservedMemory: {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		constants.FieldVirtualMachineRunStrategy: {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
			Default:  string(kubevirtv1.RunStrategyRerunOnFailure),
			ValidateFunc: validation.StringInSlice([]string{
				string(kubevirtv1.RunStrategyAlways),
				string(kubevirtv1.RunStrategyManual),
				string(kubevirtv1.RunStrategyHalted),
				string(kubevirtv1.RunStrategyRerunOnFailure),
			}, false),
		},
		constants.FieldVirtualMachineSSHKeys: {
			Type:     schema.TypeList,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		constants.FieldVirtualMachineCloudInit: {
			Type:     schema.TypeList,
			Optional: true,
			ForceNew: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: virtualmachine.ResourceCloudInitSchema(),
			},
		},
		constants.FieldVirtualMachineDisk: {
			Type:     schema.TypeList,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Resource{
				Schema: virtualmachine.ResourceDiskSchema(),
			},
		},
		constants.FieldVirtualMachineNetworkInterface: {
			Type:     schema.TypeList,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Resource{
				Schema: virtualmachine.ResourceNetworkInterfaceSchema(),
			},
		},
		constants.FieldVirtualMachineInput: {
			Type:     schema.TypeList,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Resource{
				Schema: virtualmachine.ResourceInputSchema(),
			},
		},
		constants.FieldVirtualMachineTPM: {
			Type:     schema.TypeList,
			Optional: true,
			ForceNew: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: virtualmachine.ResourceTPMSchema(),
			},
		},
		constants.FieldVirtualMachineEFI: {
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
			Default:  false,
		},
		constants.FieldVirtualMachineSecureBoot: {
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
			Default:  false,
		},
		constants.FieldVirtualMachineCPUPinning: {
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
			Default:  false,
		},
		constants.FieldVirtualMachineIsolateEmulatorThread: {
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
			Default:  false,
		},
		constants.FieldVirtualMachineNodeSelector: {
			Type:     schema.TypeMap,
			Optional: true,
			ForceNew: true,
		},
	}
	util.NamespacedSchemaWrap(s, false)
	// Override name to be optional+computed (auto-generated by Harvester if not set)
	s[constants.FieldCommonName].Required = false
	s[constants.FieldCommonName].Optional = true
	s[constants.FieldCommonName].Computed = true
	// All fields must be ForceNew since versions are immutable (no UpdateContext)
	s[constants.FieldCommonDescription].ForceNew = true
	s[constants.FieldCommonTags].ForceNew = true
	s[constants.FieldCommonLabels].ForceNew = true
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
