package kubeovn_provider_network

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func DataSourceKubeOVNProviderNetwork() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKubeOVNProviderNetworkRead,
		Schema:      DataSourceSchema(),
	}
}

func dataSourceKubeOVNProviderNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	obj, err := c.KubeOVNClient.KubeovnV1().ProviderNetworks().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(dataSourceKubeOVNProviderNetworkImport(d, obj))
}

func dataSourceKubeOVNProviderNetworkImport(d *schema.ResourceData, obj *kubeovnv1.ProviderNetwork) error {
	stateGetter, err := importer.ResourceKubeOVNProviderNetworkStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}
