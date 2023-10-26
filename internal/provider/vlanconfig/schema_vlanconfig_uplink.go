package vlanconfig

import (
	harvsternetworkv1 "github.com/harvester/harvester-network-controller/pkg/apis/network.harvesterhci.io/v1beta1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func resourceUplinkSchema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldUplinkNICs: {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		constants.FieldUplinkBondMode: {
			Type:     schema.TypeString,
			Optional: true,
			Default:  string(harvsternetworkv1.BondMoDeActiveBackup),
			ValidateFunc: validation.StringInSlice([]string{
				string(harvsternetworkv1.BondMoDeActiveBackup),
				string(harvsternetworkv1.BondMode8023AD),
				string(harvsternetworkv1.BondModeBalanceAlb),
				string(harvsternetworkv1.BondModeBalanceTlb),
				string(harvsternetworkv1.BondModeBalanceRr),
				string(harvsternetworkv1.BondModeBalanceXor),
				string(harvsternetworkv1.BondModeBroadcast),
			}, false),
		},
		constants.FieldUplinkBondMiimon: {
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntAtLeast(-1),
			Description:  "refer to https://www.kernel.org/doc/Documentation/networking/bonding.txt",
		},
		constants.FieldUplinkMTU: {
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntAtLeast(0),
		},
	}
	return s
}
