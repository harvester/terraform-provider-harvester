package network

import (
	networkutils "github.com/harvester/harvester-network-controller/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldNetworkVlanID: {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntBetween(0, 4094),
			Description:  "e.g. 0-4094",
		},
		constants.FieldNetworkConfig: {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		constants.FieldNetworkClusterNetworkName: {
			Type:         schema.TypeString,
			Required:     true,
			Description:  "", //TODO
			ValidateFunc: validation.StringIsNotEmpty,
		},
		constants.FieldNetworkRouteMode: {
			Type:     schema.TypeString,
			Optional: true,
			Default:  string(networkutils.Auto),
			ValidateFunc: validation.StringInSlice([]string{
				string(networkutils.Auto),
				string(networkutils.Manual),
			}, false),
		},
		constants.FieldNetworkRouteDHCPServerIP: {
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{constants.FieldNetworkRouteCIDR, constants.FieldNetworkRouteGateWay},
		},
		constants.FieldNetworkRouteCIDR: {
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{constants.FieldNetworkRouteDHCPServerIP},
			Description:   "e.g. 172.16.0.1/24",
		},
		constants.FieldNetworkRouteGateWay: {
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{constants.FieldNetworkRouteDHCPServerIP},
			Description:   "e.g. 172.16.0.1",
		},
		constants.FieldNetworkRouteConnectivity: {
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
