package sriovgpudevice

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	devicesv1 "github.com/harvester/pcidevices/pkg/apis/devices.harvesterhci.io/v1beta1"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func ResourceSRIOVGPUDevice() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSRIOVGPUDeviceCreate,
		ReadContext:   resourceSRIOVGPUDeviceRead,
		DeleteContext: resourceSRIOVGPUDeviceDelete,
		UpdateContext: resourceSRIOVGPUDeviceUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: Schema(),
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(5 * time.Minute),
			Read:    schema.DefaultTimeout(2 * time.Minute),
			Update:  schema.DefaultTimeout(5 * time.Minute),
			Delete:  schema.DefaultTimeout(2 * time.Minute),
			Default: schema.DefaultTimeout(2 * time.Minute),
		},
	}
}

func resourceSRIOVGPUDeviceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// SR-IOV device resources are created by the PCI Device controller. Therefore,
	// configuring an SR-IOV device with the Harvester Terraform provider is
	// always an update operation.
	return resourceSRIOVGPUDeviceUpdate(ctx, d, meta)
}

func resourceSRIOVGPUDeviceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get(constants.FieldCommonName).(string)
	obj, err := c.HarvesterDeviceClient.DevicesV1beta1().SRIOVGPUDevices().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceSRIOVGPUDeviceImport(d, obj))
}

func resourceSRIOVGPUDeviceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	err := d.Set(constants.FieldSRIOVGPUDeviceEnabled, false)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceSRIOVGPUDeviceUpdate(ctx, d, meta)
}

func resourceSRIOVGPUDeviceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	obj, err := c.HarvesterDeviceClient.DevicesV1beta1().SRIOVGPUDevices().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	toUpdate, err := util.ResourceConstruct(ctx, d, Updater(obj))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = c.HarvesterDeviceClient.DevicesV1beta1().SRIOVGPUDevices().Update(ctx, toUpdate.(*devicesv1.SRIOVGPUDevice), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(resourceSRIOVGPUDeviceWaitForState(ctx, d, meta, schema.TimeoutUpdate))
}

func resourceSRIOVGPUDeviceWaitForState(ctx context.Context, d *schema.ResourceData, meta any, timeoutKey string) error {
	var (
		pending []string
		target  []string
	)

	enabled := d.Get(constants.FieldSRIOVGPUDeviceEnabled).(bool)
	if !enabled {
		pending = []string{devicesv1.DeviceEnabled}
		target = []string{devicesv1.DeviceDisabled}
	} else {
		pending = []string{devicesv1.DeviceDisabled}
		target = []string{devicesv1.DeviceEnabled}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    resourceSRIOVGPUDeviceRefresh(ctx, d, meta),
		Timeout:    d.Timeout(timeoutKey),
		Delay:      1 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func resourceSRIOVGPUDeviceRefresh(ctx context.Context, d *schema.ResourceData, meta any) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		var state string
		c, err := meta.(*config.Config).K8sClient()
		if err != nil {
			return nil, "", err
		}
		name := d.Get(constants.FieldCommonName).(string)
		obj, err := c.HarvesterDeviceClient.DevicesV1beta1().SRIOVGPUDevices().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return obj, constants.StateCommonRemoved, nil
			}
			return obj, constants.StateCommonError, err
		}

		if err = resourceSRIOVGPUDeviceImport(d, obj); err != nil {
			return obj, constants.StateCommonError, err
		}

		vfAddresses, ok := d.GetOk(constants.FieldSRIOVGPUDeviceVFAddresses)
		if !ok || (ok && len(vfAddresses.([]any)) < 1) {
			return obj, devicesv1.DeviceDisabled, nil
		}

		numPCIDevicesReady, err := countPCIDevices(ctx, c, obj)
		if err != nil {
			return obj, constants.StateCommonError, err
		}

		stateRaw := d.Get(constants.FieldSRIOVGPUDeviceEnabled).(bool)
		if stateRaw && len(vfAddresses.([]any)) != 0 && numPCIDevicesReady == len(vfAddresses.([]any)) {
			state = devicesv1.DeviceEnabled
		} else {
			state = devicesv1.DeviceDisabled
		}
		return obj, state, err
	}
}

func countPCIDevices(ctx context.Context, client *client.Client, sriovGPUDevice *devicesv1.SRIOVGPUDevice) (int, error) {
	list, err := client.HarvesterDeviceClient.
		DevicesV1beta1().
		PCIDevices().
		List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", devicesv1.ParentSRIOVGPUDeviceLabel, sriovGPUDevice.Name),
		})
	if err != nil {
		return 0, err
	}
	numDevicesReady := len(list.Items)

	tflog.Info(ctx, fmt.Sprintf("found %d VFs", numDevicesReady))
	return numDevicesReady, err
}

func resourceSRIOVGPUDeviceImport(d *schema.ResourceData, obj *devicesv1.SRIOVGPUDevice) error {
	stateGetter, err := importer.ResourceSRIOVGPUDeviceStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}
