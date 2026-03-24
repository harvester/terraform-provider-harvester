package kubeovn_iptables_dnat_rule

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func ResourceKubeOVNIptablesDnatRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKubeOVNIptablesDnatRuleCreate,
		ReadContext:   resourceKubeOVNIptablesDnatRuleRead,
		UpdateContext: resourceKubeOVNIptablesDnatRuleUpdate,
		DeleteContext: resourceKubeOVNIptablesDnatRuleDelete,
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

func resourceKubeOVNIptablesDnatRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	toCreate, err := util.ResourceConstruct(d, Creator(name))
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeOVNClient.KubeovnV1().IptablesDnatRules().Create(ctx, toCreate.(*kubeovnv1.IptablesDnatRule), metav1.CreateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(helper.BuildID("", name))
	return diag.FromErr(resourceKubeOVNIptablesDnatRuleImport(d, obj))
}

func resourceKubeOVNIptablesDnatRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeOVNClient.KubeovnV1().IptablesDnatRules().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceKubeOVNIptablesDnatRuleImport(d, obj))
}

func resourceKubeOVNIptablesDnatRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeOVNClient.KubeovnV1().IptablesDnatRules().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	toUpdate, err := util.ResourceConstruct(d, Updater(obj))
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = c.KubeOVNClient.KubeovnV1().IptablesDnatRules().Update(ctx, toUpdate.(*kubeovnv1.IptablesDnatRule), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceKubeOVNIptablesDnatRuleRead(ctx, d, meta)
}

func resourceKubeOVNIptablesDnatRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.KubeOVNClient.KubeovnV1().IptablesDnatRules().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func resourceKubeOVNIptablesDnatRuleImport(d *schema.ResourceData, obj *kubeovnv1.IptablesDnatRule) error {
	stateGetter, err := importer.ResourceKubeOVNIptablesDnatRuleStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}
