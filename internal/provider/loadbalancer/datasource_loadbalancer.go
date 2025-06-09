package loadbalancer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func DataSourceLoadBalancer() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceLoadBalancerRead,
		Schema:      DataSourceSchema(),
	}
}

func dataSourceLoadBalancerRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace := data.Get(constants.FieldCommonNamespace).(string)
	name := data.Get(constants.FieldCommonName).(string)

	loadbalancer, err := c.HarvesterLoadbalancerClient.
		LoadbalancerV1beta1().
		LoadBalancers(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceLoadBalancerImport(data, loadbalancer))
}
