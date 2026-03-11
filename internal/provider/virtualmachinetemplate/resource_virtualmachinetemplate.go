package virtualmachinetemplate

import (
	"context"
	"time"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func ResourceVirtualMachineTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVirtualMachineTemplateCreate,
		ReadContext:   resourceVirtualMachineTemplateRead,
		UpdateContext: resourceVirtualMachineTemplateUpdate,
		DeleteContext: resourceVirtualMachineTemplateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: Schema(),
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(2 * time.Minute),
			Read:    schema.DefaultTimeout(2 * time.Minute),
			Update:  schema.DefaultTimeout(2 * time.Minute),
			Delete:  schema.DefaultTimeout(5 * time.Minute),
			Default: schema.DefaultTimeout(2 * time.Minute),
		},
	}
}

func resourceVirtualMachineTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace := d.Get(constants.FieldCommonNamespace).(string)
	name := d.Get(constants.FieldCommonName).(string)
	toCreate, err := util.ResourceConstruct(d, Creator(namespace, name))
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineTemplates(namespace).Create(ctx, toCreate.(*harvsterv1.VirtualMachineTemplate), metav1.CreateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(helper.BuildID(namespace, name))
	return diag.FromErr(resourceVirtualMachineTemplateImport(d, obj))
}

func resourceVirtualMachineTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineTemplates(namespace).Get(ctx, name, metav1.GetOptions{})
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
	_, err = c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineTemplates(namespace).Update(ctx, toUpdate.(*harvsterv1.VirtualMachineTemplate), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceVirtualMachineTemplateRead(ctx, d, meta)
}

func resourceVirtualMachineTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineTemplates(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceVirtualMachineTemplateImport(d, obj))
}

func resourceVirtualMachineTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineTemplates(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func resourceVirtualMachineTemplateImport(d *schema.ResourceData, obj *harvsterv1.VirtualMachineTemplate) error {
	stateGetter, err := importer.ResourceVirtualMachineTemplateStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}
