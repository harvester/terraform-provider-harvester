package virtualmachine

import (
	"github.com/harvester/harvester/pkg/builder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func resourceNetworkInterfaceSchema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FiledNetworkInterfaceName: {
			Type:     schema.TypeString,
			Required: true,
		},
		constants.FiledNetworkInterfaceType: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ValidateFunc: validation.StringInSlice([]string{
				builder.NetworkInterfaceTypeBridge,
				builder.NetworkInterfaceTypeMasquerade,
				"",
			}, false),
		},
		constants.FiledNetworkInterfaceModel: {
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
		constants.FiledNetworkInterfaceMACAddress: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		constants.FiledNetworkInterfaceIPAddress: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FiledNetworkInterfaceWaitForLease: {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "wait for this network interface to obtain an IP address. If a non-management network is used, this feature requires qemu-guest-agent installed and started in the VM, otherwise, VM creation will stuck until timeout",
		},
		constants.FiledNetworkInterfaceInterfaceName: {
			Type:     schema.TypeString,
			Computed: true,
		},
		constants.FiledNetworkInterfaceNetworkName: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "if the value is empty, management network is used",
		},
	}
	return s
}
