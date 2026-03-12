package sriovdevice

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
)

func ResourceSRIOVNetworkDevice() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSRIOVNetworkDeviceCreate,
		ReadContext:   resourceSRIOVNetworkDeviceRead,
		DeleteContext: resourceSRIOVNetworkDeviceDelete,
		UpdateContext: resourceSRIOVNetworkDeviceUpdate,
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

func resourceSRIOVNetworkDeviceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// SR-IOV device resources are created by the PCI Device controller. Therefore,
	// configuring an SR-IOV device with the Harvester Terraform provider is
	// always an update operation.
	return resourceSRIOVNetworkDeviceUpdate(ctx, d, meta)
}

func resourceSRIOVNetworkDeviceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	obj, err := c.HarvesterDeviceClient.DevicesV1beta1().SRIOVNetworkDevices().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceSRIOVNetworkDeviceImport(d, obj))
}

func resourceSRIOVNetworkDeviceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	err := d.Set(constants.FieldSRIOVNetworkDeviceNumVFs, 0)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceSRIOVNetworkDeviceUpdate(ctx, d, meta)
}

func resourceSRIOVNetworkDeviceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	obj, err := c.HarvesterDeviceClient.DevicesV1beta1().SRIOVNetworkDevices().Get(ctx, name, metav1.GetOptions{})
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
	_, err = c.HarvesterDeviceClient.DevicesV1beta1().SRIOVNetworkDevices().Update(ctx, toUpdate.(*devicesv1.SRIOVNetworkDevice), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSRIOVNetworkDeviceRead(ctx, d, meta)
}

func resourceSRIOVNetworkDeviceWaitForState(ctx context.Context, d *schema.ResourceData, meta interface{}, timeoutKey string) error {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{},
		Target:     []string{},
		Refresh:    resourceSRIOVNetworkDeviceRefresh(ctx, d, meta),
		Timeout:    d.Timeout(timeoutKey),
		Delay:      1 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func resourceSRIOVNetworkDeviceRefresh(ctx context.Context, d *schema.ResourceData, meta interface{}) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		var state string
		c, err := meta.(*config.Config).K8sClient()
		if err != nil {
			return nil, "", err
		}
		name := d.Get(constants.FieldCommonName).(string)
		obj, err := c.HarvesterDeviceClient.DevicesV1beta1().SRIOVNetworkDevices().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return obj, constants.StateCommonRemoved, nil
			}
			return obj, constants.StateCommonError, err
		}
		if err = resourceSRIOVNetworkDeviceImport(d, obj); err != nil {
			return obj, constants.StateCommonError, err
		}
		state_raw := d.Get(constants.FieldSRIOVNetworkDeviceEnabled).(bool)
		if state_raw {
			state = devicesv1.DeviceEnabled
		} else {
			state = devicesv1.DeviceDisabled
		}
		return obj, state, err
	}
}
