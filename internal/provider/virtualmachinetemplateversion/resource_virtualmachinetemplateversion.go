package virtualmachinetemplateversion

import (
	"context"
	"strings"
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

func ResourceVirtualMachineTemplateVersion() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVirtualMachineTemplateVersionCreate,
		ReadContext:   resourceVirtualMachineTemplateVersionRead,
		DeleteContext: resourceVirtualMachineTemplateVersionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: Schema(),
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(2 * time.Minute),
			Read:    schema.DefaultTimeout(2 * time.Minute),
			Delete:  schema.DefaultTimeout(5 * time.Minute),
			Default: schema.DefaultTimeout(2 * time.Minute),
		},
	}
}

func resourceVirtualMachineTemplateVersionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace := d.Get(constants.FieldCommonNamespace).(string)
	name := d.Get(constants.FieldCommonName).(string)
	toCreate, err := util.ResourceConstruct(d, Creator(c, ctx, namespace, name))
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineTemplateVersions(namespace).Create(ctx, toCreate.(*harvsterv1.VirtualMachineTemplateVersion), metav1.CreateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(helper.BuildID(obj.Namespace, obj.Name))
	return diag.FromErr(resourceVirtualMachineTemplateVersionImport(d, obj))
}

func resourceVirtualMachineTemplateVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineTemplateVersions(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceVirtualMachineTemplateVersionImport(d, obj))
}

func resourceVirtualMachineTemplateVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineTemplateVersions(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		// Harvester webhook denies deleting the default template version directly.
		// Delete the parent template instead, which cascade-deletes all its versions.
		if strings.Contains(err.Error(), "Cannot delete the default templateVersion") {
			templateID := d.Get(constants.FieldVirtualMachineTemplateVersionTemplateID).(string)
			templateNS, templateName, parseErr := helper.NamespacedNamePartsByDefault(templateID, namespace)
			if parseErr != nil {
				return diag.FromErr(err)
			}
			delErr := c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineTemplates(templateNS).Delete(ctx, templateName, metav1.DeleteOptions{})
			if delErr != nil && !apierrors.IsNotFound(delErr) {
				return diag.FromErr(err)
			}
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func resourceVirtualMachineTemplateVersionImport(d *schema.ResourceData, obj *harvsterv1.VirtualMachineTemplateVersion) error {
	stateGetter, err := importer.ResourceVirtualMachineTemplateVersionStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}
