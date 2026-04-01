package kubeovn_provider_network

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNProviderNetDefaultInterface: {
			Type:     schema.TypeString,
			Required: true,
		},
		constants.FieldKubeOVNProviderNetCustomInterfaces: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldKubeOVNCustomInterfaceInterface: {
						Type:     schema.TypeString,
						Required: true,
					},
					constants.FieldKubeOVNCustomInterfaceNodes: {
						Type:     schema.TypeList,
						Required: true,
						MinItems: 1,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
				},
			},
		},
		constants.FieldKubeOVNProviderNetExcludeNodes: {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		constants.FieldKubeOVNProviderNetExchangeLinkName: {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		constants.FieldKubeOVNProviderNetStatusReady: {
			Type:     schema.TypeBool,
			Computed: true,
		},
		constants.FieldKubeOVNProviderNetStatusReadyNodes: {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		constants.FieldKubeOVNProviderNetStatusNotReadyNodes: {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		constants.FieldKubeOVNProviderNetStatusVlans: {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
