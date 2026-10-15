package kubeovn_ippool

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func ResourceKubeOVNIPPool() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKubeOVNIPPoolCreate,
		ReadContext:   resourceKubeOVNIPPoolRead,
		UpdateContext: resourceKubeOVNIPPoolUpdate,
		DeleteContext: resourceKubeOVNIPPoolDelete,
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

func resourceKubeOVNIPPoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	toCreate, err := util.ResourceConstruct(ctx, d, Creator(name))
	if err != nil {
		return diag.FromErr(err)
	}
	client := c.KubeOVNClient.KubeovnV1().IPPools()
	if _, err = client.Create(ctx, toCreate.(*kubeovnv1.IPPool), metav1.CreateOptions{}); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(helper.BuildID("", name))
	err = util.WaitForReady(ctx, func(ctx context.Context) (bool, error) {
		obj, err := client.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return ippoolReady(obj), nil
	})
	if err != nil {
		return diag.Errorf("waiting for kube-ovn IP pool %s to be ready: %v", name, err)
	}
	return resourceKubeOVNIPPoolRead(ctx, d, meta)
}

// ippoolReady reports whether kube-ovn has loaded the pool into its IPAM: the
// controller sets the Ready condition once the address set and the IPAM are
// updated.
func ippoolReady(pool *kubeovnv1.IPPool) bool {
	cond := pool.Status.GetCondition(kubeovnv1.Ready)
	return cond != nil && cond.Status == corev1.ConditionTrue
}

func resourceKubeOVNIPPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeOVNClient.KubeovnV1().IPPools().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceKubeOVNIPPoolImport(d, obj))
}

func resourceKubeOVNIPPoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeOVNClient.KubeovnV1().IPPools().Get(ctx, name, metav1.GetOptions{})
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
	_, err = c.KubeOVNClient.KubeovnV1().IPPools().Update(ctx, toUpdate.(*kubeovnv1.IPPool), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceKubeOVNIPPoolRead(ctx, d, meta)
}

func resourceKubeOVNIPPoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	client := c.KubeOVNClient.KubeovnV1().IPPools()
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
		return diag.Errorf("waiting for kube-ovn IP pool %s to be deleted: %v", name, err)
	}
	d.SetId("")
	return nil
}

func resourceKubeOVNIPPoolImport(d *schema.ResourceData, obj *kubeovnv1.IPPool) error {
	stateGetter, err := importer.ResourceKubeOVNIPPoolStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}
