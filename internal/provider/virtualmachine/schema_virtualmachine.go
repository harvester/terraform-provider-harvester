package virtualmachine

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func resourceVirtualMachineFeaturesSchema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldFeatureACPI: {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "ACPI enables/disables ACPI inside the guest. Defaults to enabled.",
		},
		constants.FieldFeatureAPIC: {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "APIC settings for the guest VM.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldFeatureAPICEnabled: {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  true,
						Description: "Enabled determines if the feature should be enabled or disabled on the" +
							"guest. Defaults to true.",
					},
					constants.FieldFeatureAPICEndOfInterrupt: {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  false,
						Description: "EndOfInterrupt enables the end of interrupt notification in the guest." +
							" Defaults to false.",
					},
				},
			},
		},
		constants.FieldFeatureHyperV: {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Defaults to the machine type setting.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldFeatureHyperVEVMCS: {
						Type:     schema.TypeBool,
						Optional: true,
						Computed: true,
						Description: "EVMCS Speeds up L2 vmexits, but disables other virtualization features. " +
							"Requires vapic. Defaults to the machine type setting.",
					},
					constants.FieldFeatureHyperVFrequencies: {
						Type:     schema.TypeBool,
						Optional: true,
						Computed: true,
						Description: "Frequencies improves the TSC clock source handling for Hyper-V on KVM. " +
							"Defaults to the machine type setting.",
					},
					constants.FieldFeatureHyperVIPI: {
						Type:     schema.TypeBool,
						Optional: true,
						Computed: true,
						Description: "IPI improves performances in overcommited environments. Requires vpindex. " +
							"Defaults to the machine type setting.",
					},
					constants.FieldFeatureHyperVReenlightenment: {
						Type:     schema.TypeBool,
						Optional: true,
						Computed: true,
						Description: "Reenlightenment enables the notifications on TSC frequency changes. " +
							"Defaults to the machine type setting.",
					},
					constants.FieldFeatureHyperVRelaxed: {
						Type:     schema.TypeBool,
						Optional: true,
						Computed: true,
						Description: "Relaxed instructs the guest OS to disable watchdog timeouts. " +
							"Defaults to the machine type setting.",
					},
					constants.FieldFeatureHyperVReset: {
						Type:     schema.TypeBool,
						Optional: true,
						Computed: true,
						Description: "Reset enables Hyperv reboot/reset for the vmi. Requires synic. " +
							"Defaults to the machine type setting.",
					},
					constants.FieldFeatureHyperVRuntime: {
						Type:     schema.TypeBool,
						Optional: true,
						Computed: true,
						Description: "Runtime improves the time accounting to improve scheduling in the guest. " +
							"Defaults to the machine type setting.",
					},
					constants.FieldFeatureHyperVSpinlocks: {
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Description: "Spinlocks allows to configure the spinlock retry attempts.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								constants.FieldFeatureHyperVSpinlocksEnabled: {
									Type:     schema.TypeBool,
									Optional: true,
									Default:  true,
									Description: "Enabled determines if the feature should be enabled or disabled on " +
										"the guest. Defaults to true.",
								},
								constants.FieldFeatureHyperVSpinlocksRetries: {
									Type:     schema.TypeInt,
									Optional: true,
									Default:  4096,
									Description: "Retries indicates the number of retries. Must be a value greater or " +
										"equal 4096. Defaults to 4096.",
									ValidateFunc: validation.IntAtLeast(4096),
								},
							},
						},
					},
					constants.FieldFeatureHyperVSyNIC: {
						Type:     schema.TypeBool,
						Optional: true,
						Computed: true,
						Description: "SyNIC enables the Synthetic Interrupt Controller. " +
							"Defaults to the machine type setting.",
					},
					constants.FieldFeatureHyperVSyNICTimer: {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Description: "SyNICTimer enables Synthetic Interrupt Controller Timers, reducing CPU " +
							"load. Defaults to the machine type setting.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								constants.FieldFeatureHyperVSyNICTimerEnabled: {
									Type:     schema.TypeBool,
									Optional: true,
									Computed: true,
								},
								constants.FieldFeatureHyperVSyNICTimerDirect: {
									Type:     schema.TypeBool,
									Optional: true,
									Default:  true,
								},
							},
						},
					},
					constants.FieldFeatureHyperVTLBFlush: {
						Type:     schema.TypeBool,
						Optional: true,
						Computed: true,
						Description: "TLBFlush improves performances in overcommited environments. " +
							"Requires vpindex. Defaults to the machine type setting.",
					},
					constants.FieldFeatureHyperVVAPIC: {
						Type:     schema.TypeBool,
						Optional: true,
						Computed: true,
						Description: "VAPIC improves the paravirtualized handling of interrupts. " +
							"Defaults to the machine type setting.",
					},
					constants.FieldFeatureHyperVVendorId: {
						Type:     schema.TypeString,
						Optional: true,
						Description: "VendorID sets the hypervisor vendor id, visible to the vmi. " +
							"String up to twelve characters.",
					},
					constants.FieldFeatureHyperVVPIndex: {
						Type:     schema.TypeBool,
						Optional: true,
						Computed: true,
						Description: "VPIndex enables the Virtual Processor Index to help windows identifying " +
							"virtual processors. Defaults to the machine type setting.",
					},
				},
			},
		},
		constants.FieldFeatureHyperVPassthrough: {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
			Description: "This enables all supported hyperv flags automatically. Bear in mind that if " +
				"this enabled hyperV features cannot be enabled explicitly. In addition, a Virtual " +
				"Machine using it will be non-migratable.",
		},
		constants.FieldFeatureKVM: {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Configure how KVM presence is exposed to the guest.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldFeatureKVMHidden: {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  false,
						Description: "Hide the KVM hypervisor from standard MSR based discovery. Defaults to" +
							" false",
					},
				},
			},
		},
		constants.FieldFeaturePVSpinLock: {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
			Description: "Notify the guest that the host supports paravirtual spinlocks. For older " +
				"kernels this feature should be explicitly disabled.",
		},
		constants.FieldFeatureSMM: {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "SMM enables/disables System Management Mode. TSEG not yet implemented.",
		},
	}
	return s
}

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldVirtualMachineFeatures: {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: resourceVirtualMachineFeaturesSchema(),
			},
		},
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
		constants.FieldVirtualMachineCPUModel: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "CPU model for the virtual machine",
		},
		constants.FieldVirtualMachineMemory: {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "1Gi",
		},
		constants.FieldVirtualMachineRequests: {
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Description: "Resource requests for the VM. When unset, Harvester's overcommit webhook manages these values.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldRequestsCPU: {
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
						Description: "CPU request as Kubernetes quantity (e.g. 1, 500m).",
					},
					constants.FieldRequestsMemory: {
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
						Description: "Memory request as Kubernetes quantity (e.g. 512Mi, 1Gi).",
					},
				},
			},
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
		constants.FieldVirtualMachineNodeSelector: {
			Type:        schema.TypeMap,
			Description: "Node selector for scheduling the VM. The key is the label key and the value is the label value.",
			Optional:    true,
		},
		constants.FieldVirtualMachineCreateInitialSnapshot: {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Create an initial snapshot named {vm-name}-initial after the VM is created and ready",
		},
		constants.FieldVirtualMachineHostDevice: {
			Type:        schema.TypeList,
			Description: "Attaches a host device to the VM",
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldHostDeviceName: {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Name of the host device",
					},
					constants.FieldHostDeviceDeviceName: {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Device name (resource name) of the host device",
					},
				},
			},
		},
	}
	util.NamespacedSchemaWrap(s, false)
	s[constants.FieldCommonTags].Description = "The tag is reflected as label on the VM.\n" +
		"For example: `sample-tag = sample` adds label `tag.harvesterhci.io/sample-tag: sample`.\n" +
		"For `ssh-user` tag, the value is added to `cloudinit.user_data` if:\n" +
		"1. Both `cloudinit.user_data_base64` and `cloudinit.user_data_secret_name` are empty.\n" +
		"2. There is no `user` field in `cloudinit.user_data`.\n"
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
