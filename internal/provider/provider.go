package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/provider/clusternetwork"
	"github.com/harvester/terraform-provider-harvester/internal/provider/image"
	"github.com/harvester/terraform-provider-harvester/internal/provider/keypair"
	"github.com/harvester/terraform-provider-harvester/internal/provider/network"
	"github.com/harvester/terraform-provider-harvester/internal/provider/virtualmachine"
	"github.com/harvester/terraform-provider-harvester/internal/provider/volume"
	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			constants.FiledProviderKubeConfig: {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "harvester kubeconfig",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			constants.ResourceTypeImage:          image.DataSourceImage(),
			constants.ResourceTypeKeyPair:        keypair.DataSourceKeypair(),
			constants.ResourceTypeNetwork:        network.DataSourceNetwork(),
			constants.ResourceTypeVirtualMachine: virtualmachine.DataSourceVirtualMachine(),
			constants.ResourceTypeVolume:         volume.DataSourceVolume(),
			constants.ResourceTypeClusterNetwork: clusternetwork.DataSourceClusterNetwork(),
		},
		ResourcesMap: map[string]*schema.Resource{
			constants.ResourceTypeImage:          image.ResourceImage(),
			constants.ResourceTypeKeyPair:        keypair.ResourceKeypair(),
			constants.ResourceTypeNetwork:        network.ResourceNetwork(),
			constants.ResourceTypeVirtualMachine: virtualmachine.ResourceVirtualMachine(),
			constants.ResourceTypeVolume:         volume.ResourceVolume(),
			constants.ResourceTypeClusterNetwork: clusternetwork.ResourceClusterNetwork(),
		},
	}
	p.ConfigureContextFunc = configure(p)
	return p
}

func configure(p *schema.Provider) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		kubeConfig := d.Get(constants.FiledProviderKubeConfig).(string)
		c, err := client.NewClient(kubeConfig)
		if err != nil {
			return nil, diag.FromErr(err)
		}
		return c, nil
	}
}
