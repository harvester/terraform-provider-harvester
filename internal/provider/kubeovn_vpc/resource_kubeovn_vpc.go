package kubeovn_vpc

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
	// kube-ovn sets no condition on a VPC: status.standby turns true once
	// the controller has created its logical router.
	err = util.WaitForReady(ctx, func(ctx context.Context) (bool, error) {
		vpc, err := vpcs.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return vpc.Status.Standby, nil
	})
	if err != nil {
		return diag.Errorf("waiting for kube-ovn VPC %s to be ready: %v", name, err)
	}
	return resourceKubeOVNVpcRead(ctx, d, meta)
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
	err = util.RetryDelete(ctx, func(ctx context.Context) error {
		return vpcs.Delete(ctx, name, metav1.DeleteOptions{})
	})
	if err != nil {
		return diag.Errorf("deleting kube-ovn VPC %s: %v", name, err)
	}
	err = util.WaitForDeletion(ctx, func(ctx context.Context) error {
		_, err := vpcs.Get(ctx, name, metav1.GetOptions{})
		return err
	})
	if err != nil {
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
