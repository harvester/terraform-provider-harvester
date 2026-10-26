package kubeovn_iptables_eip

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

func ResourceKubeOVNIptablesEIP() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKubeOVNIptablesEIPCreate,
		ReadContext:   resourceKubeOVNIptablesEIPRead,
		UpdateContext: resourceKubeOVNIptablesEIPUpdate,
		DeleteContext: resourceKubeOVNIptablesEIPDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: Schema(),
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(5 * time.Minute),
			Read:    schema.DefaultTimeout(2 * time.Minute),
			Update:  schema.DefaultTimeout(2 * time.Minute),
			Delete:  schema.DefaultTimeout(2 * time.Minute),
			Default: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func resourceKubeOVNIptablesEIPCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	toCreate, err := util.ResourceConstruct(ctx, d, Creator(name))
	if err != nil {
		return diag.FromErr(err)
	}
	eips := c.KubeOVNClient.KubeovnV1().IptablesEIPs()
	if _, err = eips.Create(ctx, toCreate.(*kubeovnv1.IptablesEIP), metav1.CreateOptions{}); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(helper.BuildID("", name))
	// The EIP only becomes ready once the NAT gateway pod runs, which can
	// take minutes on a new gateway.
	stateConf := &retry.StateChangeConf{
		Pending:    []string{constants.StateCommonActive},
		Target:     []string{constants.StateCommonReady},
		Refresh:    eipRefresh(ctx, eips, name),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      1 * time.Second,
		MinTimeout: 2 * time.Second,
	}
	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("waiting for kube-ovn iptables EIP %s to be ready: %v", name, err)
	}
	return resourceKubeOVNIptablesEIPRead(ctx, d, meta)
}

// eipRefresh reports the iptables EIP as ready once the kube-ovn controller
// has applied it in the NAT gateway pod and set status.ready.
func eipRefresh(ctx context.Context, eips kubeovnclient.IptablesEIPInterface, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		obj, err := eips.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return obj, constants.StateCommonRemoved, nil
			}
			return obj, constants.StateCommonError, err
		}
		return obj, importer.KubeOVNReadyState(obj.Status.Ready), nil
	}
}

func resourceKubeOVNIptablesEIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeOVNClient.KubeovnV1().IptablesEIPs().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceKubeOVNIptablesEIPImport(d, obj))
}

func resourceKubeOVNIptablesEIPUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeOVNClient.KubeovnV1().IptablesEIPs().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	toUpdate, err := util.ResourceConstruct(ctx, d, Updater(obj))
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = c.KubeOVNClient.KubeovnV1().IptablesEIPs().Update(ctx, toUpdate.(*kubeovnv1.IptablesEIP), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceKubeOVNIptablesEIPRead(ctx, d, meta)
}

func resourceKubeOVNIptablesEIPDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	eips := c.KubeOVNClient.KubeovnV1().IptablesEIPs()
	if err = eips.Delete(ctx, name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return diag.Errorf("deleting kube-ovn iptables EIP %s: %v", name, err)
	}
	// The EIP keeps the kube-ovn finalizer until it is released, and the
	// kube-ovn webhook refuses to delete the NAT gateway before that.
	stateConf := &retry.StateChangeConf{
		Pending:    []string{constants.StateCommonActive, constants.StateCommonReady},
		Target:     []string{constants.StateCommonRemoved},
		Refresh:    eipRefresh(ctx, eips, name),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      1 * time.Second,
		MinTimeout: 2 * time.Second,
	}
	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("waiting for kube-ovn iptables EIP %s to be deleted: %v", name, err)
	}
	d.SetId("")
	return nil
}

func resourceKubeOVNIptablesEIPImport(d *schema.ResourceData, obj *kubeovnv1.IptablesEIP) error {
	stateGetter, err := importer.ResourceKubeOVNIptablesEIPStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}
