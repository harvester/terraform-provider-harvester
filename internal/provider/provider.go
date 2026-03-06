package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/go-homedir"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/provider/addon"
	"github.com/harvester/terraform-provider-harvester/internal/provider/bootstrap"
	"github.com/harvester/terraform-provider-harvester/internal/provider/cloudinitsecret"
	"github.com/harvester/terraform-provider-harvester/internal/provider/clusternetwork"
	"github.com/harvester/terraform-provider-harvester/internal/provider/image"
	"github.com/harvester/terraform-provider-harvester/internal/provider/ippool"
	"github.com/harvester/terraform-provider-harvester/internal/provider/keypair"
	kubeovnip "github.com/harvester/terraform-provider-harvester/internal/provider/kubeovn_ip"
	kubeovnippool "github.com/harvester/terraform-provider-harvester/internal/provider/kubeovn_ippool"
	kubeovndnat "github.com/harvester/terraform-provider-harvester/internal/provider/kubeovn_iptables_dnat_rule"
	kubeovneip "github.com/harvester/terraform-provider-harvester/internal/provider/kubeovn_iptables_eip"
	kubeovnfip "github.com/harvester/terraform-provider-harvester/internal/provider/kubeovn_iptables_fip_rule"
	kubeovnsnat "github.com/harvester/terraform-provider-harvester/internal/provider/kubeovn_iptables_snat_rule"
	kubeovnovndnat "github.com/harvester/terraform-provider-harvester/internal/provider/kubeovn_ovn_dnat_rule"
	kubeovnovneip "github.com/harvester/terraform-provider-harvester/internal/provider/kubeovn_ovn_eip"
	kubeovnovnfip "github.com/harvester/terraform-provider-harvester/internal/provider/kubeovn_ovn_fip"
	kubeovnovnsnat "github.com/harvester/terraform-provider-harvester/internal/provider/kubeovn_ovn_snat_rule"
	kubeovnprovnet "github.com/harvester/terraform-provider-harvester/internal/provider/kubeovn_provider_network"
	kubeovnqos "github.com/harvester/terraform-provider-harvester/internal/provider/kubeovn_qos_policy"
	kubeovnsg "github.com/harvester/terraform-provider-harvester/internal/provider/kubeovn_security_group"
	kubeovnsubnet "github.com/harvester/terraform-provider-harvester/internal/provider/kubeovn_subnet"
	kubeovnslr "github.com/harvester/terraform-provider-harvester/internal/provider/kubeovn_switch_lb_rule"
	kubeovnvip "github.com/harvester/terraform-provider-harvester/internal/provider/kubeovn_vip"
	kubeovnvlan "github.com/harvester/terraform-provider-harvester/internal/provider/kubeovn_vlan"
	kubeovnvpc "github.com/harvester/terraform-provider-harvester/internal/provider/kubeovn_vpc"
	kubeovnvpcdns "github.com/harvester/terraform-provider-harvester/internal/provider/kubeovn_vpc_dns"
	kubeovnegw "github.com/harvester/terraform-provider-harvester/internal/provider/kubeovn_vpc_egress_gateway"
	kubeovnnatgw "github.com/harvester/terraform-provider-harvester/internal/provider/kubeovn_vpc_nat_gateway"
	"github.com/harvester/terraform-provider-harvester/internal/provider/loadbalancer"
	"github.com/harvester/terraform-provider-harvester/internal/provider/network"
	"github.com/harvester/terraform-provider-harvester/internal/provider/schedulebackup"
	"github.com/harvester/terraform-provider-harvester/internal/provider/setting"
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
			constants.FieldProviderKubeConfig: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "kubeconfig file path or content of the kubeconfig file as base64 encoded string, users can use the KUBECONFIG environment variable instead.",
			},
			constants.FieldProviderKubeContext: {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "name of the kubernetes context to use",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			constants.ResourceTypeAddon:                   addon.DataSourceAddon(),
			constants.ResourceTypeCloudInitSecret:         cloudinitsecret.DataSourceCloudInitSecret(),
			constants.ResourceTypeClusterNetwork:          clusternetwork.DataSourceClusterNetwork(),
			constants.ResourceTypeIPPool:                  ippool.DataSourceIPPool(),
			constants.ResourceTypeImage:                   image.DataSourceImage(),
			constants.ResourceTypeKeyPair:                 keypair.DataSourceKeypair(),
			constants.ResourceTypeKubeOVNIP:               kubeovnip.DataSourceKubeOVNIP(),
			constants.ResourceTypeKubeOVNIptablesDnatRule: kubeovndnat.DataSourceKubeOVNIptablesDnatRule(),
			constants.ResourceTypeKubeOVNIptablesEIP:      kubeovneip.DataSourceKubeOVNIptablesEIP(),
			constants.ResourceTypeKubeOVNIptablesFIPRule:  kubeovnfip.DataSourceKubeOVNIptablesFIPRule(),
			constants.ResourceTypeKubeOVNIptablesSnatRule: kubeovnsnat.DataSourceKubeOVNIptablesSnatRule(),
			constants.ResourceTypeKubeOVNIPPool:           kubeovnippool.DataSourceKubeOVNIPPool(),
			constants.ResourceTypeKubeOVNProviderNetwork:  kubeovnprovnet.DataSourceKubeOVNProviderNetwork(),
			constants.ResourceTypeKubeOVNQoSPolicy:        kubeovnqos.DataSourceKubeOVNQoSPolicy(),
			constants.ResourceTypeKubeOVNSecurityGroup:    kubeovnsg.DataSourceKubeOVNSecurityGroup(),
			constants.ResourceTypeKubeOVNSubnet:           kubeovnsubnet.DataSourceKubeOVNSubnet(),
			constants.ResourceTypeKubeOVNVlan:             kubeovnvlan.DataSourceKubeOVNVlan(),
			constants.ResourceTypeKubeOVNVpc:              kubeovnvpc.DataSourceKubeOVNVpc(),
			constants.ResourceTypeKubeOVNOvnDnatRule:      kubeovnovndnat.DataSourceKubeOVNOvnDnatRule(),
			constants.ResourceTypeKubeOVNOvnEip:           kubeovnovneip.DataSourceKubeOVNOvnEip(),
			constants.ResourceTypeKubeOVNOvnFip:           kubeovnovnfip.DataSourceKubeOVNOvnFip(),
			constants.ResourceTypeKubeOVNOvnSnatRule:      kubeovnovnsnat.DataSourceKubeOVNOvnSnatRule(),
			constants.ResourceTypeKubeOVNSwitchLBRule:     kubeovnslr.DataSourceKubeOVNSwitchLBRule(),
			constants.ResourceTypeKubeOVNVip:              kubeovnvip.DataSourceKubeOVNVip(),
			constants.ResourceTypeKubeOVNVpcDns:           kubeovnvpcdns.DataSourceKubeOVNVpcDns(),
			constants.ResourceTypeKubeOVNVpcEgressGateway: kubeovnegw.DataSourceKubeOVNVpcEgressGateway(),
			constants.ResourceTypeKubeOVNVpcNatGateway:    kubeovnnatgw.DataSourceKubeOVNVpcNatGateway(),
			constants.ResourceTypeLoadBalancer:            loadbalancer.DataSourceLoadBalancer(),
			constants.ResourceTypeNetwork:                 network.DataSourceNetwork(),
			constants.ResourceTypeScheduleBackup:          schedulebackup.DataSourceScheduleBackup(),
			constants.ResourceTypeSetting:                 setting.DataSourceSetting(),
			constants.ResourceTypeStorageClass:            storageclass.DataSourceStorageClass(),
			constants.ResourceTypeVLANConfig:              vlanconfig.DataSourceVLANConfig(),
			constants.ResourceTypeVirtualMachine:          virtualmachine.DataSourceVirtualMachine(),
			constants.ResourceTypeVolume:                  volume.DataSourceVolume(),
		},
		ResourcesMap: map[string]*schema.Resource{
			constants.ResourceTypeAddon:                   addon.ResourceAddon(),
			constants.ResourceTypeCloudInitSecret:         cloudinitsecret.ResourceCloudInitSecret(),
			constants.ResourceTypeClusterNetwork:          clusternetwork.ResourceClusterNetwork(),
			constants.ResourceTypeIPPool:                  ippool.ResourceIPPool(),
			constants.ResourceTypeImage:                   image.ResourceImage(),
			constants.ResourceTypeKeyPair:                 keypair.ResourceKeypair(),
			constants.ResourceTypeKubeOVNIptablesDnatRule: kubeovndnat.ResourceKubeOVNIptablesDnatRule(),
			constants.ResourceTypeKubeOVNIptablesEIP:      kubeovneip.ResourceKubeOVNIptablesEIP(),
			constants.ResourceTypeKubeOVNIptablesFIPRule:  kubeovnfip.ResourceKubeOVNIptablesFIPRule(),
			constants.ResourceTypeKubeOVNIptablesSnatRule: kubeovnsnat.ResourceKubeOVNIptablesSnatRule(),
			constants.ResourceTypeKubeOVNIPPool:           kubeovnippool.ResourceKubeOVNIPPool(),
			constants.ResourceTypeKubeOVNProviderNetwork:  kubeovnprovnet.ResourceKubeOVNProviderNetwork(),
			constants.ResourceTypeKubeOVNQoSPolicy:        kubeovnqos.ResourceKubeOVNQoSPolicy(),
			constants.ResourceTypeKubeOVNSecurityGroup:    kubeovnsg.ResourceKubeOVNSecurityGroup(),
			constants.ResourceTypeKubeOVNSubnet:           kubeovnsubnet.ResourceKubeOVNSubnet(),
			constants.ResourceTypeKubeOVNVlan:             kubeovnvlan.ResourceKubeOVNVlan(),
			constants.ResourceTypeKubeOVNVpc:              kubeovnvpc.ResourceKubeOVNVpc(),
			constants.ResourceTypeKubeOVNOvnDnatRule:      kubeovnovndnat.ResourceKubeOVNOvnDnatRule(),
			constants.ResourceTypeKubeOVNOvnEip:           kubeovnovneip.ResourceKubeOVNOvnEip(),
			constants.ResourceTypeKubeOVNOvnFip:           kubeovnovnfip.ResourceKubeOVNOvnFip(),
			constants.ResourceTypeKubeOVNOvnSnatRule:      kubeovnovnsnat.ResourceKubeOVNOvnSnatRule(),
			constants.ResourceTypeKubeOVNSwitchLBRule:     kubeovnslr.ResourceKubeOVNSwitchLBRule(),
			constants.ResourceTypeKubeOVNVip:              kubeovnvip.ResourceKubeOVNVip(),
			constants.ResourceTypeKubeOVNVpcDns:           kubeovnvpcdns.ResourceKubeOVNVpcDns(),
			constants.ResourceTypeKubeOVNVpcEgressGateway: kubeovnegw.ResourceKubeOVNVpcEgressGateway(),
			constants.ResourceTypeKubeOVNVpcNatGateway:    kubeovnnatgw.ResourceKubeOVNVpcNatGateway(),
			constants.ResourceTypeLoadBalancer:            loadbalancer.ResourceLoadBalancer(),
			constants.ResourceTypeNetwork:                 network.ResourceNetwork(),
			constants.ResourceTypeScheduleBackup:          schedulebackup.ResourceScheduleBackup(),
			constants.ResourceTypeSetting:                 setting.ResourceSetting(),
			constants.ResourceTypeStorageClass:            storageclass.ResourceStorageClass(),
			constants.ResourceTypeVLANConfig:              vlanconfig.ResourceVLANConfig(),
			constants.ResourceTypeVirtualMachine:          virtualmachine.ResourceVirtualMachine(),
			constants.ResourceTypeVolume:                  volume.ResourceVolume(),
			constants.ResourceTypeBootstrap:               bootstrap.ResourceBootstrap(),
		},
		ConfigureContextFunc: providerConfig,
	}
	return p
}

func providerConfig(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	bootstrap := d.Get(constants.FieldProviderBootstrap).(bool)
	kubeConfig := d.Get(constants.FieldProviderKubeConfig).(string)
	kubeContext := d.Get(constants.FieldProviderKubeContext).(string)
	if bootstrap {
		if kubeConfig != "" {
			return nil, diag.Errorf("kubeconfig is not allowed when bootstrap is true")
		}

		if kubeContext != "" {
			return nil, diag.Errorf("kubecontext is not allowed when bootstrap is true")
		}

		return &config.Config{
			Bootstrap: bootstrap,
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
