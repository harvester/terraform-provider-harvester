package setting

import (
	"context"
	"encoding/json"
	"reflect"
	"slices"
	"time"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	harvclient "github.com/harvester/harvester/pkg/generated/clientset/versioned"
	"github.com/harvester/harvester/pkg/settings"
	"github.com/harvester/harvester/pkg/util/network"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/util"
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
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	obj, err := c.HarvesterClient.HarvesterhciV1beta1().Settings().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return updateSetting(ctx, c.HarvesterClient, d, obj)
}

func resourceSettingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
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
	return updateSetting(ctx, c.HarvesterClient, d, obj)
}

func resourceSettingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
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
	handleStorageNetworkSetting(obj, d.Get(constants.FieldSettingValue).(string))
	return diag.FromErr(resourceSettingImport(d, obj))
}

// The setting cannot be deleted; it can only be reset to an empty string (""), which represents using the default value.
func resourceSettingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
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

// handleStorageNetworkSetting handles the special case for the "storage-network" setting.
// If the new value is functionally equivalent to the old value (ignoring the order of excluded networks),
// it retains the old value to avoid unnecessary updates.
func handleStorageNetworkSetting(newSetting *harvsterv1.Setting, oldValue string) {
	if newSetting.Name != settings.StorageNetworkName {
		return
	}

	var (
		oldConfig network.Config
		newConfig network.Config
		err       error
	)
	if err = json.Unmarshal([]byte(oldValue), &oldConfig); err != nil {
		logrus.WithError(err).
			WithField("value", oldValue).
			Warn("Failed to unmarshal old storage-network setting")
		return
	}
	if err = json.Unmarshal([]byte(newSetting.Value), &newConfig); err != nil {
		logrus.WithError(err).
			WithField("value", newSetting.Value).
			Warn("Failed to unmarshal new storage-network setting")
		return
	}

	slices.Sort(oldConfig.Exclude)
	slices.Sort(newConfig.Exclude)
	if reflect.DeepEqual(oldConfig, newConfig) {
		newSetting.Value = oldValue
	}
}

func updateSetting(ctx context.Context, harvesterClient *harvclient.Clientset, d *schema.ResourceData, oldSetting *harvsterv1.Setting) diag.Diagnostics {
	oldValue := oldSetting.Value

	toUpdate, err := util.ResourceConstruct(d, Updater(oldSetting))
	if err != nil {
		return diag.FromErr(err)
	}

	newSetting := toUpdate.(*harvsterv1.Setting)
	newValue := newSetting.Value
	handleStorageNetworkSetting(newSetting, oldValue)

	newSetting, err = harvesterClient.HarvesterhciV1beta1().Settings().Update(ctx, toUpdate.(*harvsterv1.Setting), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	newSetting.Value = newValue
	return diag.FromErr(resourceSettingImport(d, newSetting))
}
