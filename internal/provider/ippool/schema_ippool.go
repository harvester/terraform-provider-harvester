package ippool

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldIPPoolDescription: {
			Type:     schema.TypeString,
			Optional: true,
		},
		constants.SubresourceTypeIPPoolRange: {
			Type:     schema.TypeList,
			Required: true,
			MinItems: 1,
			Elem: &schema.Resource{
				Schema: subresourceSchemaIPPoolRange(),
			},
			Description: "IP Range belonging to this pool, can be given multiple times",
		},
		constants.SubresourceTypeIPPoolSelector: {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: subresourceSchemaIPPoolSelector(),
			},
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}

func subresourceSchemaIPPoolRange() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldRangeStart: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "",
		},
		constants.FieldRangeEnd: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "",
		},
		constants.FieldRangeSubnet: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "",
		},
		constants.FieldRangeGateway: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "",
		},
	}
	return s
}

func subresourceSchemaIPPoolSelector() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldSelectorPriority: {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Priority of the IP pool. Large numbers have higher priority",
		},
		constants.FieldSelectorNetwork: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Namespace/name of the VM network",
		},
		constants.FieldSelectorScope: {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Scope of the IP pool",
			Elem: &schema.Resource{
				Schema: subresourceSchemaIPPoolSelectorScope(),
			},
		},
	}
	return s
}

func subresourceSchemaIPPoolSelectorScope() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldScopeProject: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Name of the project",
		},
		constants.FieldScopeNamespace: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Namespace of the VMs of the guest cluster",
		},
		constants.FieldScopeGuestCluster: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Name of the guest cluster",
		},
	}
	return s
}
