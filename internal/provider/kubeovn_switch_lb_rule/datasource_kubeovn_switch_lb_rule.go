package kubeovn_switch_lb_rule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func DataSourceKubeOVNSwitchLBRule() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKubeOVNSwitchLBRuleRead,
		Schema:      DataSourceSchema(),
	}
}

func dataSourceKubeOVNSwitchLBRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	obj, err := c.KubeOVNClient.KubeovnV1().SwitchLBRules().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	stateGetter, err := importer.ResourceKubeOVNSwitchLBRuleStateGetter(obj)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(util.ResourceStatesSet(d, stateGetter))
}
