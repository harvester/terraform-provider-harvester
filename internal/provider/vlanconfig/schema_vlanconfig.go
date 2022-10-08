package vlanconfig

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldVLANConfigClusterNetworkName: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringNotInSlice([]string{constants.ManagementClusterNetworkName}, true),
			Description:  "mgmt is a built-in cluster network and does not support creating/updating network configs.",
		},
		constants.FieldVLANConfigNodeSelector: {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "refer to https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector",
		},
		constants.FieldVLANConfigMatchedNodes: {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		constants.FieldVLANConfigUplink: {
			Type:     schema.TypeList,
			Required: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: resourceUplinkSchema(),
			},
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	s := util.DataSourceSchemaWrap(Schema())
	return s
}
