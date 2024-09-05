package virtualmachine

import (
	"context"
	"strings"
	"time"

	harvesterutil "github.com/harvester/harvester/pkg/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"

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
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(2 * time.Minute),
			Read:    schema.DefaultTimeout(2 * time.Minute),
			Update:  schema.DefaultTimeout(2 * time.Minute),
			Delete:  schema.DefaultTimeout(5 * time.Minute),
			Default: schema.DefaultTimeout(2 * time.Minute),
		},
	}
}

func resourceVirtualMachineCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	namespace := d.Get(constants.FieldCommonNamespace).(string)
	name := d.Get(constants.FieldCommonName).(string)
	toCreate, err := util.ResourceConstruct(d, Creator(c, ctx, namespace, name))
	if err != nil {
		return diag.FromErr(err)
	}
	vm, err := c.HarvesterClient.KubevirtV1().VirtualMachines(namespace).Create(ctx, toCreate.(*kubevirtv1.VirtualMachine), metav1.CreateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(helper.BuildID(namespace, name))
	if err = updateLocalFields(d, constants.FieldVirtualMachineRestartAfterUpdate); err != nil {
		return diag.FromErr(err)
	}
	runStrategy, err := vm.RunStrategy()
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceVirtualMachineWaitForState(ctx, d, meta, runStrategy, namespace, name, schema.TimeoutCreate, ""))
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
	toUpdate, err := util.ResourceConstruct(d, Updater(c, ctx, obj))
	if err != nil {
		return diag.FromErr(err)
	}
	vm, err := c.HarvesterClient.KubevirtV1().VirtualMachines(namespace).Update(ctx, toUpdate.(*kubevirtv1.VirtualMachine), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	if err = updateLocalFields(d, constants.FieldVirtualMachineRestartAfterUpdate); err != nil {
		return diag.FromErr(err)
	}
	runStrategy, err := vm.RunStrategy()
	if err != nil {
		return diag.FromErr(err)
	}
	oldInstanceUID := ""
	if IsNeedRestart(d, runStrategy) {
		vmi, err := c.HarvesterClient.KubevirtV1().VirtualMachineInstances(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return diag.FromErr(err)
		}
		if vmi != nil {
			oldInstanceUID = string(vmi.UID)
		}
		if err = c.KubeVirtSubresourceClient.Put().Namespace(namespace).Resource(constants.ResourceVirtualMachine).SubResource(constants.SubresourceRestart).Name(name).Do(ctx).Error(); err != nil {
			return diag.FromErr(err)
		}
	}
	return diag.FromErr(resourceVirtualMachineWaitForState(ctx, d, meta, runStrategy, namespace, name, schema.TimeoutUpdate, oldInstanceUID))
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
		if !apierrors.IsNotFound(err) {
			return diag.FromErr(err)
		}
		vmi = nil
	}
	return diag.FromErr(resourceVirtualMachineImport(d, vm, vmi, ""))
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
	if diskList, ok := d.GetOk(constants.FieldVirtualMachineDisk); ok {
		for _, disk := range diskList.([]interface{}) {
			r := disk.(map[string]interface{})
			diskName := r[constants.FieldDiskName].(string)
			deleteConfigs[diskName] = r[constants.FieldDiskAutoDelete].(bool)
		}
	}
	removedPVCs := make([]string, 0, len(vm.Spec.Template.Spec.Volumes))
	for _, volume := range vm.Spec.Template.Spec.Volumes {
		if volume.PersistentVolumeClaim == nil {
			continue
		}
		if autoDelete, ok := deleteConfigs[volume.Name]; ok && !autoDelete {
			continue
		}
		removedPVCs = append(removedPVCs, volume.PersistentVolumeClaim.ClaimName)
	}
	vmCopy := vm.DeepCopy()
	vmCopy.Annotations[harvesterutil.RemovedPVCsAnnotationKey] = strings.Join(removedPVCs, ",")
	_, err = c.HarvesterClient.KubevirtV1().VirtualMachines(namespace).Update(ctx, vmCopy, metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	propagationPolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{PropagationPolicy: &propagationPolicy}
	if err = c.HarvesterClient.KubevirtV1().VirtualMachines(namespace).Delete(ctx, name, deleteOptions); err != nil && !apierrors.IsNotFound(err) {
		return diag.FromErr(err)
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			constants.StateCommonReady,
			constants.StateCommonFailed,
			constants.StateCommonUnknown,
			constants.StateVirtualMachineRunning,
			constants.StateVirtualMachineStarting,
			constants.StateVirtualMachineStopping,
			constants.StateVirtualMachineStopped,
		},
		Target:     []string{constants.StateCommonRemoved},
		Refresh:    resourceVirtualMachineRefresh(ctx, d, meta, namespace, name, ""),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceVirtualMachineImport(d *schema.ResourceData, vm *kubevirtv1.VirtualMachine, vmi *kubevirtv1.VirtualMachineInstance, oldInstanceUID string) error {
	stateGetter, err := importer.ResourceVirtualMachineStateGetter(vm, vmi, oldInstanceUID)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}

func resourceVirtualMachineWaitForState(ctx context.Context, d *schema.ResourceData, meta interface{}, runStrategy kubevirtv1.VirtualMachineRunStrategy, namespace, name, timeOutKey, oldInstanceUID string) error {
	var (
		pending = []string{constants.StateVirtualMachineStarting, constants.StateVirtualMachineStopping, constants.StateVirtualMachineRunning, constants.StateCommonFailed, constants.StateCommonUnknown}
		target  []string
	)
	switch runStrategy {
	case kubevirtv1.RunStrategyHalted:
		pending = append(pending, constants.StateCommonReady)
		target = []string{constants.StateVirtualMachineStopped}
	case kubevirtv1.RunStrategyAlways, kubevirtv1.RunStrategyRerunOnFailure:
		pending = append(pending, constants.StateVirtualMachineStopped)
		target = []string{constants.StateCommonReady}
	default:
		return nil
	}
	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    resourceVirtualMachineRefresh(ctx, d, meta, namespace, name, oldInstanceUID),
		Timeout:    d.Timeout(timeOutKey),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func resourceVirtualMachineRefresh(ctx context.Context, d *schema.ResourceData, meta interface{}, namespace, name, oldInstanceUID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		c := meta.(*client.Client)
		vm, err := c.HarvesterClient.KubevirtV1().VirtualMachines(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return vm, constants.StateCommonRemoved, nil
			}
			return vm, constants.StateCommonError, err
		}
		vmi, err := c.HarvesterClient.KubevirtV1().VirtualMachineInstances(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return vm, constants.StateCommonError, err
			}
			vmi = nil
		}
		if err = resourceVirtualMachineImport(d, vm, vmi, oldInstanceUID); err != nil {
			return vm, constants.StateCommonError, err
		}
		state := d.Get(constants.FieldCommonState).(string)
		return vm, state, nil
	}
}

func updateLocalFields(d *schema.ResourceData, keys ...string) error {
	for _, key := range keys {
		if err := d.Set(key, d.Get(key)); err != nil {
			return err
		}
	}
	return nil
}

func IsNeedRestart(d *schema.ResourceData, runStrategy kubevirtv1.VirtualMachineRunStrategy) bool {
	switch runStrategy {
	case kubevirtv1.RunStrategyAlways, kubevirtv1.RunStrategyRerunOnFailure:
		return d.Get(constants.FieldVirtualMachineRestartAfterUpdate).(bool)
	}
	return false
}
