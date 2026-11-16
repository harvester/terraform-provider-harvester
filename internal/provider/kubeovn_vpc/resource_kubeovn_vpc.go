package kubeovn_vpc

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

func ResourceKubeOVNVpc() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKubeOVNVpcCreate,
		ReadContext:   resourceKubeOVNVpcRead,
		UpdateContext: resourceKubeOVNVpcUpdate,
		DeleteContext: resourceKubeOVNVpcDelete,
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

func resourceKubeOVNVpcCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	toCreate, err := util.ResourceConstruct(ctx, d, Creator(name))
	if err != nil {
		return diag.FromErr(err)
	}
	vpcs := c.KubeOVNClient.KubeovnV1().Vpcs()
	if _, err = vpcs.Create(ctx, toCreate.(*kubeovnv1.Vpc), metav1.CreateOptions{}); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(helper.BuildID("", name))
	stateConf := &retry.StateChangeConf{
		Pending:    []string{constants.StateCommonActive},
		Target:     []string{constants.StateCommonReady},
		Refresh:    vpcRefresh(ctx, vpcs, name),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      1 * time.Second,
		MinTimeout: 2 * time.Second,
	}
	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("waiting for kube-ovn VPC %s to be ready: %v", name, err)
	}
	return resourceKubeOVNVpcRead(ctx, d, meta)
}

// vpcRefresh reports the VPC as ready once kube-ovn has created its logical
// router: the controller sets no condition on a VPC, only status.standby.
func vpcRefresh(ctx context.Context, vpcs kubeovnclient.VpcInterface, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		vpc, err := vpcs.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return vpc, constants.StateCommonRemoved, nil
			}
			return vpc, constants.StateCommonError, err
		}
		return vpc, importer.KubeOVNReadyState(vpc.Status.Standby), nil
	}
}

func resourceKubeOVNVpcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeOVNClient.KubeovnV1().Vpcs().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceKubeOVNVpcImport(d, obj))
}

func resourceKubeOVNVpcUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeOVNClient.KubeovnV1().Vpcs().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	toUpdate, err := util.ResourceConstruct(ctx, d, Updater(obj))
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = c.KubeOVNClient.KubeovnV1().Vpcs().Update(ctx, toUpdate.(*kubeovnv1.Vpc), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceKubeOVNVpcRead(ctx, d, meta)
}

func resourceKubeOVNVpcDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	vpcs := c.KubeOVNClient.KubeovnV1().Vpcs()
	// The kube-ovn webhook refuses to delete a VPC while it still has
	// subnets, which may still be finalizing: retry until the delete is
	// accepted, then wait for the VPC finalizer to be released.
	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		if err := vpcs.Delete(ctx, name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			return retry.RetryableError(err)
		}
		return nil
	})
	if err != nil {
		return diag.Errorf("deleting kube-ovn VPC %s: %v", name, err)
	}
	stateConf := &retry.StateChangeConf{
		Pending:    []string{constants.StateCommonActive, constants.StateCommonReady},
		Target:     []string{constants.StateCommonRemoved},
		Refresh:    vpcRefresh(ctx, vpcs, name),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      1 * time.Second,
		MinTimeout: 2 * time.Second,
	}
	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("waiting for kube-ovn VPC %s to be deleted: %v", name, err)
	}
	d.SetId("")
	return nil
}

func resourceKubeOVNVpcImport(d *schema.ResourceData, obj *kubeovnv1.Vpc) error {
	stateGetter, err := importer.ResourceKubeOVNVpcStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}
