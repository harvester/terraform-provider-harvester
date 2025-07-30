package vlanconfig

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func DataSourceVLANConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVLANConfigRead,
		Schema:      DataSourceSchema(),
	}
}

func dataSourceVLANConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)

	vlanConfig, err := c.HarvesterNetworkClient.NetworkV1beta1().VlanConfigs().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceVLANConfigImport(d, vlanConfig))
}
