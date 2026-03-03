package image

import (
	"context"
	"errors"
	"fmt"
	"time"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func ResourceImage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceImageCreate,
		ReadContext:   resourceImageRead,
		DeleteContext: resourceImageDelete,
		UpdateContext: resourceImageUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: Schema(),
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(60 * time.Minute),
			Read:    schema.DefaultTimeout(2 * time.Minute),
			Update:  schema.DefaultTimeout(30 * time.Minute),
			Delete:  schema.DefaultTimeout(10 * time.Minute),
			Default: schema.DefaultTimeout(2 * time.Minute),
		},
	}
}

func resourceImageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	_, err = c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(namespace).Create(ctx, toCreate.(*harvsterv1.VirtualMachineImage), metav1.CreateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(helper.BuildID(namespace, name))
	return diag.FromErr(resourceImageWaitForState(ctx, d, meta, schema.TimeoutCreate))
}

func resourceImageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(namespace).Get(ctx, name, metav1.GetOptions{})
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
	_, err = c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(namespace).Update(ctx, toUpdate.(*harvsterv1.VirtualMachineImage), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceImageRead(ctx, d, meta)
}

func resourceImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceImageImport(d, obj))
}

func resourceImageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return diag.FromErr(err)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{constants.StateImageTerminating, constants.StateCommonActive},
		Target:     []string{constants.StateCommonRemoved},
		Refresh:    resourceImageRefresh(ctx, d, meta),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      1 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceImageImport(d *schema.ResourceData, obj *harvsterv1.VirtualMachineImage) error {
	stateGetter, err := importer.ResourceImageStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}

func resourceImageWaitForState(ctx context.Context, d *schema.ResourceData, meta interface{}, timeOutKey string) error {
	lastProgress := -1
	lastProgressTime := time.Now()

	staleTimeout := 5 * time.Minute
	hardTimeout := 2 * time.Hour
	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if time.Since(startTime) > hardTimeout {
			return errors.New("hard timeout reached during image creation")
		}

		refresh := resourceImageRefresh(ctx, d, meta)
		_, state, err := refresh()
		if err != nil {
			return err
		}

		if state == constants.StateCommonActive {
			return nil
		}

		// Fail early if state is Error or Failed
		if state == constants.StateCommonError || state == constants.StateCommonFailed {
			message := d.Get(constants.FieldCommonMessage).(string)
			return errors.New(message)
		}

		// Check progress
		progress := d.Get(constants.FieldImageProgress).(int)
		if progress > lastProgress {
			lastProgress = progress
			lastProgressTime = time.Now()
		} else if time.Since(lastProgressTime) > staleTimeout {
			return fmt.Errorf("image creation stalled at %d%% for more than %v", progress, staleTimeout)
		}

		time.Sleep(15 * time.Second)
	}
}

func resourceImageRefresh(ctx context.Context, d *schema.ResourceData, meta interface{}) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		c, err := meta.(*config.Config).K8sClient()
		if err != nil {
			return nil, "", err
		}
		namespace := d.Get(constants.FieldCommonNamespace).(string)
		name := d.Get(constants.FieldCommonName).(string)
		obj, err := c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return obj, constants.StateCommonRemoved, nil
			}
			return obj, constants.StateCommonError, err
		}
		if err = resourceImageImport(d, obj); err != nil {
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
