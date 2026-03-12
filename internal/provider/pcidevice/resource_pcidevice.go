package pcidevice

import (
	"context"
	"errors"
	"time"

	devicesv1 "github.com/harvester/pcidevices/pkg/apis/devices.harvesterhci.io/v1beta1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func ResourcePCIDevice() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePCIDeviceCreate,
		ReadContext:   resourcePCIDeviceRead,
		DeleteContext: resourcePCIDeviceDelete,
		UpdateContext: resourcePCIDeviceUpdate,
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

func resourcePCIDeviceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	enabled := d.Get(constants.FieldPCIDevicePassthroughEnabled).(bool)
	obj, err := c.HarvesterDeviceClient.DevicesV1beta1().PCIDevices().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	_, err = c.HarvesterDeviceClient.DevicesV1beta1().PCIDeviceClaims().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			if enabled {
				toCreate := &devicesv1.PCIDeviceClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: name,
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "devices.harvesterhci.io/v1beta1",
								Kind:       "PCIDevice",
								Name:       name,
								UID:        obj.UID,
							},
						},
					},
					Spec: devicesv1.PCIDeviceClaimSpec{
						Address:  obj.Status.Address,
						NodeName: obj.Status.NodeName,
						UserName: "admin",
					},
				}
				_, err = c.HarvesterDeviceClient.DevicesV1beta1().PCIDeviceClaims().Create(ctx, toCreate, metav1.CreateOptions{})
				if err != nil {
					return diag.FromErr(err)
				}
			}
			return diag.FromErr(resourcePCIDeviceImport(d, obj))
		}
		return diag.FromErr(err)
	}

	return diag.FromErr(resourcePCIDeviceImport(d, obj))
}

func resourcePCIDeviceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	obj, err := c.HarvesterDeviceClient.DevicesV1beta1().PCIDevices().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	_, err = c.HarvesterDeviceClient.DevicesV1beta1().PCIDeviceClaims().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			err = d.Set(constants.FieldPCIDevicePassthroughEnabled, false)
			if err != nil {
				return diag.FromErr(err)
			}
			return diag.FromErr(resourcePCIDeviceImport(d, obj))
		}
		return diag.FromErr(err)
	}

	err = d.Set(constants.FieldPCIDevicePassthroughEnabled, true)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(resourcePCIDeviceImport(d, obj))
}

func resourcePCIDeviceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	err = c.HarvesterDeviceClient.DevicesV1beta1().PCIDeviceClaims().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourcePCIDeviceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	enabled := d.Get(constants.FieldPCIDevicePassthroughEnabled).(bool)

	_, err = c.HarvesterDeviceClient.DevicesV1beta1().PCIDeviceClaims().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			if enabled {
				return resourcePCIDeviceCreate(ctx, d, meta)
			}
			return nil
		}
		return diag.FromErr(err)
	}

	if !enabled {
		return resourcePCIDeviceDelete(ctx, d, meta)
	}

	return resourcePCIDeviceRead(ctx, d, meta)
}

func resourcePCIDeviceWaitForState(ctx context.Context, d *schema.ResourceData, meta interface{}, timeoutKey string) error {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{},
		Target:     []string{},
		Refresh:    resourcePCIDeviceRefresh(ctx, d, meta),
		Timeout:    d.Timeout(timeoutKey),
		Delay:      1 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func resourcePCIDeviceRefresh(ctx context.Context, d *schema.ResourceData, meta interface{}) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		c, err := meta.(*config.Config).K8sClient()
		if err != nil {
			return nil, "", err
		}
		name := d.Get(constants.FieldCommonName).(string)
		obj, err := c.HarvesterDeviceClient.DevicesV1beta1().PCIDevices().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return obj, constants.StateCommonRemoved, nil
			}
			return obj, constants.StateCommonError, err
		}
		if err = resourcePCIDeviceImport(d, obj); err != nil {
			return obj, constants.StateCommonError, err
		}
		state := d.Get(constants.FieldCommonState).(string)
		if state == constants.StateCommonFailed {
			message := d.Get(constants.FieldCommonMessage).(string)
			return obj, state, errors.New(message)
		}
		return obj, state, err
	}
}
