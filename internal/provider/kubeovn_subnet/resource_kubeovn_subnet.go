package kubeovn_subnet

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	kubeovnclient "github.com/kubeovn/kube-ovn/pkg/client/clientset/versioned/typed/kubeovn/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func ResourceKubeOVNSubnet() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKubeOVNSubnetCreate,
		ReadContext:   resourceKubeOVNSubnetRead,
		UpdateContext: resourceKubeOVNSubnetUpdate,
		DeleteContext: resourceKubeOVNSubnetDelete,
		CustomizeDiff: resourceKubeOVNSubnetCustomizeDiff,
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

func resourceKubeOVNSubnetCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	return planExcludeIPs(d)
}

func resourceKubeOVNSubnetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	toCreate, err := util.ResourceConstruct(ctx, d, Creator(name))
	if err != nil {
		return diag.FromErr(err)
	}
	subnets := c.KubeOVNClient.KubeovnV1().Subnets()
	if _, err = subnets.Create(ctx, toCreate.(*kubeovnv1.Subnet), metav1.CreateOptions{}); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(helper.BuildID("", name))
	stateConf := &retry.StateChangeConf{
		Pending:    []string{constants.StateCommonActive},
		Target:     []string{constants.StateCommonReady},
		Refresh:    subnetRefresh(ctx, subnets, name),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      1 * time.Second,
		MinTimeout: 2 * time.Second,
	}
	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("waiting for kube-ovn subnet %s to be ready: %v", name, err)
	}
	return resourceKubeOVNSubnetRead(ctx, d, meta)
}

// ErrSubnetRejected is returned when the kube-ovn controller marks the spec
// as not validated: the subnet cannot become ready until the spec changes.
var ErrSubnetRejected = errors.New("kube-ovn rejected the subnet")

// subnetReady reports whether kube-ovn has set up the subnet.
func subnetReady(subnet *kubeovnv1.Subnet) (bool, error) {
	if subnet.Status.IsReady() {
		return true, nil
	}
	if cond := subnet.Status.GetCondition(kubeovnv1.Validated); cond != nil && cond.Status == corev1.ConditionFalse {
		return false, fmt.Errorf("%w: %s", ErrSubnetRejected, cond.Message)
	}
	return false, nil
}

func subnetRefresh(ctx context.Context, subnets kubeovnclient.SubnetInterface, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		subnet, err := subnets.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return subnet, constants.StateCommonRemoved, nil
			}
			return subnet, constants.StateCommonError, err
		}
		ready, err := subnetReady(subnet)
		switch {
		case err != nil:
			return subnet, constants.StateCommonError, err
		case ready:
			return subnet, constants.StateCommonReady, nil
		}
		return subnet, constants.StateCommonActive, nil
	}
}

func resourceKubeOVNSubnetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeOVNClient.KubeovnV1().Subnets().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceKubeOVNSubnetImport(d, obj))
}

func resourceKubeOVNSubnetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeOVNClient.KubeovnV1().Subnets().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	toUpdate, err := util.ResourceConstruct(ctx, d, Updater(obj))
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = c.KubeOVNClient.KubeovnV1().Subnets().Update(ctx, toUpdate.(*kubeovnv1.Subnet), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceKubeOVNSubnetRead(ctx, d, meta)
}

func resourceKubeOVNSubnetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	subnets := c.KubeOVNClient.KubeovnV1().Subnets()
	// The kube-ovn webhook refuses to delete a subnet while IPs are still
	// allocated in it, which happens while the pods of a VM or a NAT gateway
	// deleted by the same destroy are terminating: retry until it accepts.
	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		if err := subnets.Delete(ctx, name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			return retry.RetryableError(err)
		}
		return nil
	})
	if err != nil {
		return diag.Errorf("deleting kube-ovn subnet %s: %v", name, err)
	}
	// The subnet keeps the kube-ovn finalizer until its logical switch is
	// gone, and the VPC cannot be deleted before that.
	stateConf := &retry.StateChangeConf{
		Pending:    []string{constants.StateCommonActive, constants.StateCommonReady},
		Target:     []string{constants.StateCommonRemoved},
		Refresh:    subnetRefresh(ctx, subnets, name),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      1 * time.Second,
		MinTimeout: 2 * time.Second,
	}
	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("waiting for kube-ovn subnet %s to be deleted: %v", name, err)
	}
	d.SetId("")
	return nil
}

func resourceKubeOVNSubnetImport(d *schema.ResourceData, obj *kubeovnv1.Subnet) error {
	stateGetter, err := importer.ResourceKubeOVNSubnetStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}
