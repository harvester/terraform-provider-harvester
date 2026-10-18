package kubeovn_vlan

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		// kube-ovn retags the localnet port of every subnet on the VLAN when
		// the ID changes, so the change is applied in place.
		constants.FieldKubeOVNVlanID: {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntBetween(constants.KubeOVNVlanIDMin, constants.KubeOVNVlanIDMax),
			Description:  "802.1Q tag of the VLAN, 0 for untagged traffic. Changing it retags the subnets of the VLAN in place.",
		},
		// The CRD requires spec.provider. kube-ovn only retags the subnets on
		// update and never moves them to another provider network.
		constants.FieldKubeOVNVlanProvider: {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Name of the kube-ovn provider network that carries the VLAN.",
		},
		constants.FieldKubeOVNVlanStatusSubnets: {
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Names of the kube-ovn subnets attached to the VLAN.",
		},
		constants.FieldKubeOVNVlanStatusConflict: {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Whether another VLAN uses the same vlan_id on the same provider network. kube-ovn accepts such a VLAN but refuses to set up the subnets attached to it.",
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
