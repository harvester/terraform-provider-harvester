package blockdevice

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldBlockDeviceNodeName: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "name of the node to which the block device is attached",
		},
		constants.FieldBlockDeviceDevPath: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "the device path of the disk, e.g. /dev/sda",
		},
		constants.FieldBlockDeviceProvision: {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "whether the device should be provisioned for storage",
		},
		constants.FieldBlockDeviceForceFormatted: {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "force format the device to overwrite existing filesystem",
		},
		constants.FieldBlockDeviceDeviceTags: {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "device tags for provisioner, e.g. [\"default\", \"small\", \"ssd\"]",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		constants.FieldBlockDeviceProvisioner: {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "provisioner configuration for the block device",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldBlockDeviceProvisionerLonghorn: {
						Type:          schema.TypeList,
						Optional:      true,
						MaxItems:      1,
						Description:   "Longhorn volume backend disk provisioner",
						ConflictsWith: []string{constants.FieldBlockDeviceProvisioner + ".0." + constants.FieldBlockDeviceProvisionerLVM},
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								constants.FieldBlockDeviceProvisionerLonghornEV: {
									Type:         schema.TypeString,
									Optional:     true,
									Default:      "LonghornV2",
									ValidateFunc: validation.StringInSlice([]string{"LonghornV1", "LonghornV2"}, false),
									Description:  "engine version: LonghornV1 or LonghornV2",
								},
								constants.FieldBlockDeviceProvisionerLonghornDD: {
									Type:         schema.TypeString,
									Optional:     true,
									Default:      "auto",
									ValidateFunc: validation.StringInSlice([]string{"", "auto", "aio"}, false),
									Description:  "disk driver for V2 data engine: auto or aio",
								},
							},
						},
					},
					constants.FieldBlockDeviceProvisionerLVM: {
						Type:          schema.TypeList,
						Optional:      true,
						MaxItems:      1,
						Description:   "LVM volume backend disk provisioner",
						ConflictsWith: []string{constants.FieldBlockDeviceProvisioner + ".0." + constants.FieldBlockDeviceProvisionerLonghorn},
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								constants.FieldBlockDeviceProvisionerLVMVGName: {
									Type:        schema.TypeString,
									Required:    true,
									Description: "volume group name",
								},
								constants.FieldBlockDeviceProvisionerLVMParameters: {
									Type:        schema.TypeList,
									Optional:    true,
									Description: "additional LVM parameters",
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
							},
						},
					},
				},
			},
		},

		// Status fields (Computed)
		constants.FieldBlockDeviceProvisionPhase: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "provisioning phase: Provisioned, Unprovisioned, or Unprovisioning",
		},
		constants.FieldBlockDeviceDeviceStatus: {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "device hardware and filesystem status",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldBlockDeviceStatusDevPath: {
						Type:     schema.TypeString,
						Computed: true,
					},
					constants.FieldBlockDeviceStatusParentDevice: {
						Type:     schema.TypeString,
						Computed: true,
					},
					constants.FieldBlockDeviceStatusPartitioned: {
						Type:     schema.TypeBool,
						Computed: true,
					},
					constants.FieldBlockDeviceStatusCapacitySizeBytes: {
						Type:     schema.TypeInt,
						Computed: true,
					},
					constants.FieldBlockDeviceStatusDeviceType: {
						Type:     schema.TypeString,
						Computed: true,
					},
					constants.FieldBlockDeviceStatusDriveType: {
						Type:     schema.TypeString,
						Computed: true,
					},
					constants.FieldBlockDeviceStatusStorageController: {
						Type:     schema.TypeString,
						Computed: true,
					},
					constants.FieldBlockDeviceStatusVendor: {
						Type:     schema.TypeString,
						Computed: true,
					},
					constants.FieldBlockDeviceStatusModel: {
						Type:     schema.TypeString,
						Computed: true,
					},
					constants.FieldBlockDeviceStatusSerialNumber: {
						Type:     schema.TypeString,
						Computed: true,
					},
					constants.FieldBlockDeviceStatusWWN: {
						Type:     schema.TypeString,
						Computed: true,
					},
					constants.FieldBlockDeviceStatusBusPath: {
						Type:     schema.TypeString,
						Computed: true,
					},
					constants.FieldBlockDeviceStatusFSType: {
						Type:     schema.TypeString,
						Computed: true,
					},
					constants.FieldBlockDeviceStatusMountPoint: {
						Type:     schema.TypeString,
						Computed: true,
					},
					constants.FieldBlockDeviceStatusIsReadOnly: {
						Type:     schema.TypeBool,
						Computed: true,
					},
					constants.FieldBlockDeviceStatusIsRemovable: {
						Type:     schema.TypeBool,
						Computed: true,
					},
				},
			},
		},
	}
	util.NamespacedSchemaWrap(s, false)
	s[constants.FieldCommonNamespace].Default = constants.NamespaceLonghornSystem
	// Labels are managed by NDM, mark as computed
	s[constants.FieldCommonLabels].Computed = true
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
