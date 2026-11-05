package kubeovn_iptables_snat_rule

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	kubeovnclient "github.com/kubeovn/kube-ovn/pkg/client/clientset/versioned/typed/kubeovn/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func ResourceKubeOVNIptablesSnatRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKubeOVNIptablesSnatRuleCreate,
		ReadContext:   resourceKubeOVNIptablesSnatRuleRead,
		UpdateContext: resourceKubeOVNIptablesSnatRuleUpdate,
		DeleteContext: resourceKubeOVNIptablesSnatRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: Schema(),
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(2 * time.Minute),
			Read:    schema.DefaultTimeout(2 * time.Minute),
			Update:  schema.DefaultTimeout(2 * time.Minute),
			Delete:  schema.DefaultTimeout(2 * time.Minute),
			Default: schema.DefaultTimeout(2 * time.Minute),
		},
	}
}

func resourceKubeOVNIptablesSnatRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	toCreate, err := util.ResourceConstruct(ctx, d, Creator(name))
	if err != nil {
		return diag.FromErr(err)
	}
	rules := c.KubeOVNClient.KubeovnV1().IptablesSnatRules()
	if _, err = rules.Create(ctx, toCreate.(*kubeovnv1.IptablesSnatRule), metav1.CreateOptions{}); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(helper.BuildID("", name))
	stateConf := &retry.StateChangeConf{
		Pending:    []string{constants.StateCommonActive},
		Target:     []string{constants.StateCommonReady},
		Refresh:    snatRuleRefresh(ctx, rules, name),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      1 * time.Second,
		MinTimeout: 2 * time.Second,
	}
	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("waiting for kube-ovn iptables SNAT rule %s to be ready: %v", name, err)
	}
	return resourceKubeOVNIptablesSnatRuleRead(ctx, d, meta)
}

// snatRuleRefresh reports the iptables SNAT rule as ready once the kube-ovn controller
// has applied it in the NAT gateway pod and set status.ready.
func snatRuleRefresh(ctx context.Context, rules kubeovnclient.IptablesSnatRuleInterface, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		obj, err := rules.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return obj, constants.StateCommonRemoved, nil
			}
			return obj, constants.StateCommonError, err
		}
		return obj, importer.KubeOVNReadyState(obj.Status.Ready), nil
	}
}

func resourceKubeOVNIptablesSnatRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeOVNClient.KubeovnV1().IptablesSnatRules().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceKubeOVNIptablesSnatRuleImport(d, obj))
}

func resourceKubeOVNIptablesSnatRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeOVNClient.KubeovnV1().IptablesSnatRules().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	toUpdate, err := util.ResourceConstruct(ctx, d, Updater(obj))
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = c.KubeOVNClient.KubeovnV1().IptablesSnatRules().Update(ctx, toUpdate.(*kubeovnv1.IptablesSnatRule), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceKubeOVNIptablesSnatRuleRead(ctx, d, meta)
}

func resourceKubeOVNIptablesSnatRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	rules := c.KubeOVNClient.KubeovnV1().IptablesSnatRules()
	if err = rules.Delete(ctx, name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return diag.Errorf("deleting kube-ovn iptables SNAT rule %s: %v", name, err)
	}
	// The rule keeps the kube-ovn finalizer until it is removed from the NAT
	// gateway, and the kube-ovn webhook refuses to delete the EIP it
	// references before that.
	stateConf := &retry.StateChangeConf{
		Pending:    []string{constants.StateCommonActive, constants.StateCommonReady},
		Target:     []string{constants.StateCommonRemoved},
		Refresh:    snatRuleRefresh(ctx, rules, name),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      1 * time.Second,
		MinTimeout: 2 * time.Second,
	}
	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("waiting for kube-ovn iptables SNAT rule %s to be deleted: %v", name, err)
	}
	d.SetId("")
	return nil
}

func resourceKubeOVNIptablesSnatRuleImport(d *schema.ResourceData, obj *kubeovnv1.IptablesSnatRule) error {
	stateGetter, err := importer.ResourceKubeOVNIptablesSnatRuleStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}
