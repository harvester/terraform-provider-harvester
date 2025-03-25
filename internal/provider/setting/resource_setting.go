package setting

import (
	"context"
	"time"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func ResourceSetting() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSettingCreate,
		ReadContext:   resourceSettingRead,
		DeleteContext: resourceSettingDelete,
		UpdateContext: resourceSettingUpdate,
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

// The setting cannot be created. It can only be updated.
func resourceSettingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	name := d.Get(constants.FieldCommonName).(string)
	obj, err := c.HarvesterClient.HarvesterhciV1beta1().Settings().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	toUpdate, err := util.ResourceConstruct(d, Updater(obj))
	if err != nil {
		return diag.FromErr(err)
	}

	obj, err = c.HarvesterClient.HarvesterhciV1beta1().Settings().Update(ctx, toUpdate.(*harvsterv1.Setting), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(resourceSettingImport(d, obj))
}

func resourceSettingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	obj, err := c.HarvesterClient.HarvesterhciV1beta1().Settings().Get(ctx, name, metav1.GetOptions{})
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
	_, err = c.HarvesterClient.HarvesterhciV1beta1().Settings().Update(ctx, toUpdate.(*harvsterv1.Setting), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSettingRead(ctx, d, meta)
}

func resourceSettingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterClient.HarvesterhciV1beta1().Settings().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceSettingImport(d, obj))
}

// The setting cannot be deleted. It can only be resetted to empty.
func resourceSettingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	obj, err := c.HarvesterClient.HarvesterhciV1beta1().Settings().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	objCopy := obj.DeepCopy()
	objCopy.Value = ""
	_, err = c.HarvesterClient.HarvesterhciV1beta1().Settings().Update(ctx, objCopy, metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceSettingImport(d *schema.ResourceData, obj *harvsterv1.Setting) error {
	stateGetter, err := importer.ResourceSettingStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}
