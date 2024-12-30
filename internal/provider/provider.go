package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/go-homedir"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/provider/bootstrap"
	"github.com/harvester/terraform-provider-harvester/internal/provider/cloudinitsecret"
	"github.com/harvester/terraform-provider-harvester/internal/provider/clusternetwork"
	"github.com/harvester/terraform-provider-harvester/internal/provider/image"
	"github.com/harvester/terraform-provider-harvester/internal/provider/keypair"
	"github.com/harvester/terraform-provider-harvester/internal/provider/network"
	"github.com/harvester/terraform-provider-harvester/internal/provider/storageclass"
	"github.com/harvester/terraform-provider-harvester/internal/provider/virtualmachine"
	"github.com/harvester/terraform-provider-harvester/internal/provider/vlanconfig"
	"github.com/harvester/terraform-provider-harvester/internal/provider/volume"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			constants.FieldProviderBootstrap: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "bootstrap harvester server, it will write content to kubeconfig file",
			},
			constants.FieldProviderAPIUrl: {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "harvester server api url",
			},
			constants.FieldProviderKubeConfig: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "kubeconfig file path, users can use the KUBECONFIG environment variable instead",
			},
			constants.FieldProviderKubeContext: {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "name of the kubernetes context to use",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			constants.ResourceTypeImage:           image.DataSourceImage(),
			constants.ResourceTypeKeyPair:         keypair.DataSourceKeypair(),
			constants.ResourceTypeNetwork:         network.DataSourceNetwork(),
			constants.ResourceTypeVirtualMachine:  virtualmachine.DataSourceVirtualMachine(),
			constants.ResourceTypeVolume:          volume.DataSourceVolume(),
			constants.ResourceTypeClusterNetwork:  clusternetwork.DataSourceClusterNetwork(),
			constants.ResourceTypeStorageClass:    storageclass.DataSourceStorageClass(),
			constants.ResourceTypeVLANConfig:      vlanconfig.DataSourceVLANConfig(),
			constants.ResourceTypeCloudInitSecret: cloudinitsecret.DataSourceCloudInitSecret(),
		},
		ResourcesMap: map[string]*schema.Resource{
			constants.ResourceTypeImage:           image.ResourceImage(),
			constants.ResourceTypeKeyPair:         keypair.ResourceKeypair(),
			constants.ResourceTypeNetwork:         network.ResourceNetwork(),
			constants.ResourceTypeVirtualMachine:  virtualmachine.ResourceVirtualMachine(),
			constants.ResourceTypeVolume:          volume.ResourceVolume(),
			constants.ResourceTypeClusterNetwork:  clusternetwork.ResourceClusterNetwork(),
			constants.ResourceTypeStorageClass:    storageclass.ResourceStorageClass(),
			constants.ResourceTypeVLANConfig:      vlanconfig.ResourceVLANConfig(),
			constants.ResourceTypeCloudInitSecret: cloudinitsecret.ResourceCloudInitSecret(),
			constants.ResourceTypeBootstrap:       bootstrap.ResourceBootstrap(),
		},
		ConfigureContextFunc: providerConfig,
	}
	return p
}

func providerConfig(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	bootstrap := d.Get(constants.FieldProviderBootstrap).(bool)
	apiURL := d.Get(constants.FieldProviderAPIUrl).(string)
	kubeConfig := d.Get(constants.FieldProviderKubeConfig).(string)
	kubeContext := d.Get(constants.FieldProviderKubeContext).(string)
	if bootstrap {
		if apiURL == "" {
			return nil, diag.FromErr(fmt.Errorf("api url is required when bootstrap is true"))
		}

		u, err := url.Parse(apiURL)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		if kubeConfig != "" {
			return nil, diag.Errorf("kubeconfig is not allowed when bootstrap is true")
		}

		if kubeContext != "" {
			return nil, diag.Errorf("kubecontext is not allowed when bootstrap is true")
		}

		return &config.Config{
			Bootstrap: bootstrap,
			APIURL:    u.String(),
		}, nil
	}

	kubeConfig, err := homedir.Expand(d.Get(constants.FieldProviderKubeConfig).(string))
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return &config.Config{
		KubeConfig:  kubeConfig,
		KubeContext: kubeContext,
	}, nil
}
