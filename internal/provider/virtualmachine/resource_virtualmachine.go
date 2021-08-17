package virtualmachine

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/client-go/api/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func ResourceVirtualMachine() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVirtualMachineCreate,
		ReadContext:   resourceVirtualMachineRead,
		DeleteContext: resourceVirtualMachineDelete,
		UpdateContext: resourceVirtualMachineUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: Schema(),
	}
}

func resourceVirtualMachineCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	namespace := d.Get(constants.FieldCommonNamespace).(string)
	name := d.Get(constants.FieldCommonName).(string)
	toCreate, err := util.ResourceConstruct(d, Creator(namespace, name))
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterClient.KubevirtV1().VirtualMachines(namespace).Create(ctx, toCreate.(*kubevirtv1.VirtualMachine), metav1.CreateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceVirtualMachineImport(d, obj, nil)
}

func resourceVirtualMachineUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterClient.KubevirtV1().VirtualMachines(namespace).Get(ctx, name, metav1.GetOptions{})
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
	_, err = c.HarvesterClient.KubevirtV1().VirtualMachines(namespace).Update(ctx, toUpdate.(*kubevirtv1.VirtualMachine), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceVirtualMachineRead(ctx, d, meta)
}

func resourceVirtualMachineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	vm, err := c.HarvesterClient.KubevirtV1().VirtualMachines(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	vmi, err := c.HarvesterClient.KubevirtV1().VirtualMachineInstances(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return diag.FromErr(err)
	}
	return resourceVirtualMachineImport(d, vm, vmi)
}

func resourceVirtualMachineDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	vm, err := c.HarvesterClient.KubevirtV1().VirtualMachines(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	deleteConfigs := make(map[string]bool)
	if diskList, ok := d.GetOk("disk"); ok {
		for _, disk := range diskList.([]interface{}) {
			r := disk.(map[string]interface{})
			if volumeName := r["volume_name"].(string); volumeName != "" {
				deleteConfigs[volumeName] = r["auto_delete"].(bool)
			}
		}
	}
	if err = c.HarvesterClient.KubevirtV1().VirtualMachines(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return diag.FromErr(err)
	}
	for _, volume := range vm.Spec.Template.Spec.Volumes {
		if volume.DataVolume == nil {
			continue
		}
		if autoDelete, ok := deleteConfigs[volume.Name]; ok && !autoDelete {
			continue
		}
		err = c.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, volume.Name, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return diag.FromErr(err)
		}
	}
	d.SetId("")
	return nil
}

func resourceVirtualMachineImport(d *schema.ResourceData, vm *kubevirtv1.VirtualMachine, vmi *kubevirtv1.VirtualMachineInstance) diag.Diagnostics {
	stateGetter, err := importer.ResourceVirtualMachineStateGetter(vm, vmi)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(util.ResourceStatesSet(d, stateGetter))
}
