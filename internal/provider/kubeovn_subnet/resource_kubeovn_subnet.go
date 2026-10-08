package kubeovn_subnet

import (
	"context"
	"fmt"
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
	err = util.WaitForReady(ctx, func(ctx context.Context) (bool, error) {
		subnet, err := subnets.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return subnetReady(subnet)
	})
	if err != nil {
		return diag.Errorf("waiting for kube-ovn subnet %s to be ready: %v", name, err)
	}
	return resourceKubeOVNSubnetRead(ctx, d, meta)
}

// subnetReady reports whether kube-ovn has set up the subnet. It fails as
// soon as the controller marks the spec as not validated: the subnet cannot
// become ready until the spec changes.
func subnetReady(subnet *kubeovnv1.Subnet) (bool, error) {
	if subnet.Status.IsReady() {
		return true, nil
	}
	if cond := subnet.Status.GetCondition(kubeovnv1.Validated); cond != nil && cond.Status == corev1.ConditionFalse {
		return false, fmt.Errorf("kube-ovn rejected the subnet: %s", cond.Message)
	}
	return false, nil
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
	err = util.RetryDelete(ctx, func(ctx context.Context) error {
		return subnets.Delete(ctx, name, metav1.DeleteOptions{})
	})
	if err != nil {
		return diag.Errorf("deleting kube-ovn subnet %s: %v", name, err)
	}
	// The subnet keeps the kube-ovn finalizer until its logical switch is
	// gone, and the VPC cannot be deleted before that.
	err = util.WaitForDeletion(ctx, func(ctx context.Context) error {
		_, err := subnets.Get(ctx, name, metav1.GetOptions{})
		return err
	})
	if err != nil {
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
