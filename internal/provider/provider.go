package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/go-homedir"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/provider/cloudinitsecret"
	"github.com/harvester/terraform-provider-harvester/internal/provider/clusternetwork"
	"github.com/harvester/terraform-provider-harvester/internal/provider/image"
	"github.com/harvester/terraform-provider-harvester/internal/provider/ippool"
	"github.com/harvester/terraform-provider-harvester/internal/provider/keypair"
	"github.com/harvester/terraform-provider-harvester/internal/provider/loadbalancer"
	"github.com/harvester/terraform-provider-harvester/internal/provider/network"
	"github.com/harvester/terraform-provider-harvester/internal/provider/storageclass"
	"github.com/harvester/terraform-provider-harvester/internal/provider/virtualmachine"
	"github.com/harvester/terraform-provider-harvester/internal/provider/vlanconfig"
	"github.com/harvester/terraform-provider-harvester/internal/provider/volume"
	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			constants.FieldProviderKubeConfig: {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
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
			constants.ResourceTypeCloudInitSecret: cloudinitsecret.DataSourceCloudInitSecret(),
			constants.ResourceTypeClusterNetwork:  clusternetwork.DataSourceClusterNetwork(),
			constants.ResourceTypeImage:           image.DataSourceImage(),
			constants.ResourceTypeIPPool:          ippool.DataSourceIPPool(),
			constants.ResourceTypeKeyPair:         keypair.DataSourceKeypair(),
			constants.ResourceTypeLoadBalancer:    loadbalancer.DataSourceLoadBalancer(),
			constants.ResourceTypeNetwork:         network.DataSourceNetwork(),
			constants.ResourceTypeStorageClass:    storageclass.DataSourceStorageClass(),
			constants.ResourceTypeVLANConfig:      vlanconfig.DataSourceVLANConfig(),
			constants.ResourceTypeVirtualMachine:  virtualmachine.DataSourceVirtualMachine(),
			constants.ResourceTypeVolume:          volume.DataSourceVolume(),
		},
		ResourcesMap: map[string]*schema.Resource{
			constants.ResourceTypeCloudInitSecret: cloudinitsecret.ResourceCloudInitSecret(),
			constants.ResourceTypeClusterNetwork:  clusternetwork.ResourceClusterNetwork(),
			constants.ResourceTypeImage:           image.ResourceImage(),
			constants.ResourceTypeIPPool:          ippool.ResourceIPPool(),
			constants.ResourceTypeKeyPair:         keypair.ResourceKeypair(),
			constants.ResourceTypeLoadBalancer:    loadbalancer.ResourceLoadBalancer(),
			constants.ResourceTypeNetwork:         network.ResourceNetwork(),
			constants.ResourceTypeStorageClass:    storageclass.ResourceStorageClass(),
			constants.ResourceTypeVLANConfig:      vlanconfig.ResourceVLANConfig(),
			constants.ResourceTypeVirtualMachine:  virtualmachine.ResourceVirtualMachine(),
			constants.ResourceTypeVolume:          volume.ResourceVolume(),
		},
		ConfigureContextFunc: providerConfig,
	}
	return p
}

func providerConfig(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	kubeContext := d.Get(constants.FieldProviderKubeContext).(string)
	kubeConfig, err := homedir.Expand(d.Get(constants.FieldProviderKubeConfig).(string))
	if err != nil {
		return nil, diag.FromErr(err)
	}

	c, err := client.NewClient(kubeConfig, kubeContext)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	// check harvester version from settings
	serverVersion, err := c.HarvesterClient.
		HarvesterhciV1beta1().
		Settings().
		Get(ctx, "server-version", metav1.GetOptions{})
	if err != nil {
		return nil, diag.FromErr(err)
	}
	// harvester version v1.0-head, v1.0.2, v1.0.3 is not supported
	if strings.HasPrefix(serverVersion.Value, "v1.0") {
		return nil, diag.FromErr(fmt.Errorf("current Harvester server version is %s, the minimum supported version is v1.1.0", serverVersion.Value))
	}

	return c, nil
}
