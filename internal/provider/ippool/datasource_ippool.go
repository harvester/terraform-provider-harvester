package ippool

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func DataSourceIPPool() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIPPoolRead,
		Schema:      DataSourceSchema(),
	}
}

func dataSourceIPPoolRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := data.Get(constants.FieldCommonName).(string)

	ippool, err := c.HarvesterLoadbalancerClient.
		LoadbalancerV1beta1().
		IPPools().
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceIPPoolImport(data, ippool))
}
