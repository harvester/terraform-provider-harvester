package virtualmachine

import (
	"github.com/harvester/harvester/pkg/builder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func resourceNetworkInterfaceSchema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldNetworkInterfaceName: {
			Type:     schema.TypeString,
			Required: true,
		},
		constants.FieldNetworkInterfaceType: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ValidateFunc: validation.StringInSlice([]string{
				builder.NetworkInterfaceTypeBridge,
				builder.NetworkInterfaceTypeMasquerade,
				"",
			}, false),
		},
		constants.FieldNetworkInterfaceModel: {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "virtio",
			ValidateFunc: validation.StringInSlice([]string{
				"virtio",
				"e1000",
				"e1000e",
				"ne2k_pco",
				"pcnet",
				"rtl8139",
			}, false),
		},
		constants.FieldNetworkInterfaceMACAddress: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		constants.FieldNetworkInterfaceIPAddress: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldNetworkInterfaceWaitForLease: {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "wait for this network interface to obtain an IP address. If a non-management network is used, this feature requires qemu-guest-agent installed and started in the VM, otherwise, VM creation will stuck until timeout",
		},
		constants.FieldNetworkInterfaceInterfaceName: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FieldNetworkInterfaceNetworkName: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "if the value is empty, management network is used",
		},
		constants.FieldNetworkInterfaceBootOrder: {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			Description: "Boot order priority of this network interface",
		},
	}
	return s
}
