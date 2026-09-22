package vgpudevice

import (
	"context"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	devicesv1 "github.com/harvester/pcidevices/pkg/apis/devices.harvesterhci.io/v1beta1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func ResourceVGPUDevice() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVGPUDeviceCreate,
		ReadContext:   resourceVGPUDeviceRead,
		DeleteContext: resourceVGPUDeviceDelete,
		UpdateContext: resourceVGPUDeviceUpdate,
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

func resourceVGPUDeviceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return resourceVGPUDeviceUpdate(ctx, d, meta)
}

func resourceVGPUDeviceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get(constants.FieldCommonName).(string)
	obj, err := c.HarvesterDeviceClient.DevicesV1beta1().VGPUDevices().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceVGPUDeviceImport(d, obj))
}

func resourceVGPUDeviceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	obj, err := c.HarvesterDeviceClient.DevicesV1beta1().VGPUDevices().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	toUpdate, err := util.ResourceConstruct(ctx, d, Updater(obj))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = c.HarvesterDeviceClient.DevicesV1beta1().VGPUDevices().Update(ctx, toUpdate.(*devicesv1.VGPUDevice), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(resourceVGPUDeviceWaitForState(ctx, d, meta, schema.TimeoutUpdate))
}

func resourceVGPUDeviceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	err := d.Set(constants.FieldVGPUDeviceEnabled, false)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceVGPUDeviceUpdate(ctx, d, meta)
}

func resourceVGPUDeviceWaitForState(ctx context.Context, d *schema.ResourceData, meta any, timeoutKey string) error {
	var (
		pending []string
		target  []string
	)

	enabled := d.Get(constants.FieldVGPUDeviceEnabled).(bool)
	if !enabled {
		pending = []string{string(devicesv1.VGPUEnabled)}
		target = []string{string(devicesv1.VGPUDisabled)}
	} else {
		pending = []string{string(devicesv1.VGPUDisabled)}
		target = []string{string(devicesv1.VGPUEnabled)}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    resourceVGPUDeviceRefresh(ctx, d, meta),
		Timeout:    d.Timeout(timeoutKey),
		Delay:      1 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func resourceVGPUDeviceRefresh(ctx context.Context, d *schema.ResourceData, meta any) retry.StateRefreshFunc {
	return func() (any, string, error) {
		var state string
		c, err := meta.(*config.Config).K8sClient()
		if err != nil {
			return nil, "", err
		}
		name := d.Get(constants.FieldCommonName).(string)
		obj, err := c.HarvesterDeviceClient.DevicesV1beta1().VGPUDevices().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return obj, constants.StateCommonRemoved, nil
			}
			return obj, constants.StateCommonError, err
		}

		if err = resourceVGPUDeviceImport(d, obj); err != nil {
			return obj, constants.StateCommonError, err
		}

		state = d.Get(constants.FieldVGPUDeviceStatus).(string)
		return obj, state, err
	}
}

func resourceVGPUDeviceImport(d *schema.ResourceData, obj *devicesv1.VGPUDevice) error {
	stateGetter, err := importer.ResourceVGPUDeviceStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}
