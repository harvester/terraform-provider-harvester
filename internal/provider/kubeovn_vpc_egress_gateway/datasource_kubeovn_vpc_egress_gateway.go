package kubeovn_vpc_egress_gateway

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func DataSourceKubeOVNVpcEgressGateway() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKubeOVNVpcEgressGatewayRead,
		Schema:      DataSourceSchema(),
		Timeouts: &schema.ResourceTimeout{
			Read:    schema.DefaultTimeout(2 * time.Minute),
			Default: schema.DefaultTimeout(2 * time.Minute),
		},
	}
}

func dataSourceKubeOVNVpcEgressGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace := d.Get(constants.FieldCommonNamespace).(string)
	name := d.Get(constants.FieldCommonName).(string)
	obj, err := c.KubeOVNClient.KubeovnV1().VpcEgressGateways(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(helper.BuildID(namespace, name))
	stateGetter, err := importer.ResourceKubeOVNVpcEgressGatewayStateGetter(obj)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(util.ResourceStatesSet(d, stateGetter))
}
