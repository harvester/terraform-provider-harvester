package kubeovn_provider_network

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Schema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		constants.FieldKubeOVNProviderNetDefaultInterface: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Name of the host interface that kube-ovn attaches to the provider network bridge on every node, unless custom_interfaces names another one for that node.",
		},
		constants.FieldKubeOVNProviderNetCustomInterfaces: {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Interfaces to use instead of default_interface on some nodes.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					constants.FieldKubeOVNCustomInterfaceInterface: {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Name of the host interface to use on the listed nodes.",
					},
					constants.FieldKubeOVNCustomInterfaceNodes: {
						Type:        schema.TypeList,
						Required:    true,
						MinItems:    1,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Description: "Names of the nodes that use this interface.",
					},
				},
			},
		},
		constants.FieldKubeOVNProviderNetExcludeNodes: {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Names of the nodes where kube-ovn does not set up the provider network.",
		},
		constants.FieldKubeOVNProviderNetExchangeLinkName: {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Swap the names of the bridge and of the host interface, so that the bridge carries the name of the interface.",
		},
		constants.FieldKubeOVNProviderNetStatusReady: {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Whether kube-ovn has set up the provider network on every node that is not excluded.",
		},
		constants.FieldKubeOVNProviderNetStatusReadyNodes: {
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Nodes where the provider network is set up.",
		},
		constants.FieldKubeOVNProviderNetStatusNotReadyNodes: {
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Nodes where kube-ovn failed to set up the provider network, for instance because the interface does not exist there.",
		},
		constants.FieldKubeOVNProviderNetStatusVlans: {
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "Names of the kube-ovn VLANs attached to the provider network.",
		},
	}
	util.NonNamespacedSchemaWrap(s)
	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return util.DataSourceSchemaWrap(Schema())
}
