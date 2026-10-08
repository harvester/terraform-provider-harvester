package kubeovn_qos_policy

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

func ResourceKubeOVNQoSPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKubeOVNQoSPolicyCreate,
		ReadContext:   resourceKubeOVNQoSPolicyRead,
		UpdateContext: resourceKubeOVNQoSPolicyUpdate,
		DeleteContext: resourceKubeOVNQoSPolicyDelete,
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

func resourceKubeOVNQoSPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	toCreate, err := util.ResourceConstruct(ctx, d, Creator(name))
	if err != nil {
		return diag.FromErr(err)
	}
	client := c.KubeOVNClient.KubeovnV1().QoSPolicies()
	if _, err = client.Create(ctx, toCreate.(*kubeovnv1.QoSPolicy), metav1.CreateOptions{}); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(helper.BuildID("", name))
	err = util.WaitForReady(ctx, func(ctx context.Context) (bool, error) {
		obj, err := client.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return qosPolicyReady(obj), nil
	})
	if err != nil {
		return diag.Errorf("waiting for kube-ovn QoS policy %s to be ready: %v", name, err)
	}
	return resourceKubeOVNQoSPolicyRead(ctx, d, meta)
}

// qosPolicyReady reports whether kube-ovn has validated the policy: the
// controller copies shared, bindingType and the rules to the status only
// after validateQosPolicy succeeds.
func qosPolicyReady(qos *kubeovnv1.QoSPolicy) bool {
	return qos.Status.BindingType != "" && qos.Status.BindingType == qos.Spec.BindingType
}

func resourceKubeOVNQoSPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeOVNClient.KubeovnV1().QoSPolicies().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceKubeOVNQoSPolicyImport(d, obj))
}

func resourceKubeOVNQoSPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeOVNClient.KubeovnV1().QoSPolicies().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	toUpdate, err := util.ResourceConstruct(ctx, d, Updater(obj))
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = c.KubeOVNClient.KubeovnV1().QoSPolicies().Update(ctx, toUpdate.(*kubeovnv1.QoSPolicy), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceKubeOVNQoSPolicyRead(ctx, d, meta)
}

func resourceKubeOVNQoSPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	client := c.KubeOVNClient.KubeovnV1().QoSPolicies()
	err = client.Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return diag.FromErr(err)
	}
	// The kube-ovn finalizer is only released once nothing uses the object
	// any more: wait for it so that a dependent deletion does not start early.
	err = util.WaitForDeletion(ctx, func(ctx context.Context) error {
		_, err := client.Get(ctx, name, metav1.GetOptions{})
		return err
	})
	if err != nil {
		return diag.Errorf("waiting for kube-ovn QoS policy %s to be deleted: %v", name, err)
	}
	d.SetId("")
	return nil
}

func resourceKubeOVNQoSPolicyImport(d *schema.ResourceData, obj *kubeovnv1.QoSPolicy) error {
	stateGetter, err := importer.ResourceKubeOVNQoSPolicyStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}
