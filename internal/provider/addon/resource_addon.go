package addon

import (
	"context"
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
	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func ResourceAddon() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAddonCreate,
		ReadContext:   resourceAddonRead,
		UpdateContext: resourceAddonUpdate,
		DeleteContext: resourceAddonDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: Schema(),
		// Enabling or disabling an addon runs a Helm install/uninstall in the
		// cluster; heavy addons (e.g. rancher-monitoring) can take several
		// minutes. The waits below consume these timeouts, which remain
		// overridable with a `timeouts` block.
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(10 * time.Minute),
			Read:    schema.DefaultTimeout(2 * time.Minute),
			Update:  schema.DefaultTimeout(10 * time.Minute),
			Delete:  schema.DefaultTimeout(10 * time.Minute),
			Default: schema.DefaultTimeout(2 * time.Minute),
		},
	}
}

// addonCurrentState maps an addon to a retry.StateChangeConf state. It fails
// fast when the last operation failed, surfacing the controller's message.
func addonCurrentState(obj *harvsterv1.Addon) (string, error) {
	if harvsterv1.AddonOperationFailed.IsTrue(obj) {
		return string(obj.Status.Status), fmt.Errorf("addon %s/%s operation failed: %s",
			obj.Namespace, obj.Name, harvsterv1.AddonOperationFailed.GetMessage(obj))
	}
	if obj.Status.Status == harvsterv1.AddonInitState {
		return constants.StateAddonInit, nil
	}
	return string(obj.Status.Status), nil
}

func addonStateRefresh(ctx context.Context, c *client.Client, namespace, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		obj, err := c.HarvesterClient.HarvesterhciV1beta1().Addons(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return obj, constants.StateCommonRemoved, nil
			}
			return obj, constants.StateCommonError, err
		}
		state, err := addonCurrentState(obj)
		return obj, state, err
	}
}

// waitForAddonApplied waits until the addon converges to the state matching
// the configured `enabled` value: deployed when enabling, disabled when the
// addon is managed with enabled=false.
func waitForAddonApplied(ctx context.Context, c *client.Client, d *schema.ResourceData, namespace, name, timeoutKey string) error {
	var pending, target []string
	if d.Get(constants.FieldAddonEnabled).(bool) {
		pending = []string{
			constants.StateAddonInit,
			string(harvsterv1.AddonDisabled),
			string(harvsterv1.AddonDisabling),
			string(harvsterv1.AddonEnabling),
			string(harvsterv1.AddonUpdating),
		}
		target = []string{string(harvsterv1.AddonDeployed)}
	} else {
		pending = []string{
			string(harvsterv1.AddonDeployed),
			string(harvsterv1.AddonEnabling),
			string(harvsterv1.AddonUpdating),
			string(harvsterv1.AddonDisabling),
		}
		target = []string{string(harvsterv1.AddonDisabled), constants.StateAddonInit}
	}
	return waitForAddonState(ctx, c, d, namespace, name, timeoutKey, pending, target)
}

// waitForAddonDisabled waits for the disable operation performed on destroy.
func waitForAddonDisabled(ctx context.Context, c *client.Client, d *schema.ResourceData, namespace, name, timeoutKey string) error {
	pending := []string{
		string(harvsterv1.AddonDeployed),
		string(harvsterv1.AddonEnabling),
		string(harvsterv1.AddonUpdating),
		string(harvsterv1.AddonDisabling),
	}
	target := []string{string(harvsterv1.AddonDisabled), constants.StateAddonInit}
	return waitForAddonState(ctx, c, d, namespace, name, timeoutKey, pending, target)
}

func waitForAddonState(ctx context.Context, c *client.Client, d *schema.ResourceData, namespace, name, timeoutKey string, pending, target []string) error {
	stateConf := &retry.StateChangeConf{
		Pending: pending,
		Target:  target,
		Refresh: addonStateRefresh(ctx, c, namespace, name),
		Timeout: d.Timeout(timeoutKey),
		// The controller clears a stale OperationFailed condition when a new
		// operation starts; the delay leaves it room to do so before the
		// first poll.
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

// The addon cannot be created. It can only be updated (enabled/configured).
func resourceAddonCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace := d.Get(constants.FieldCommonNamespace).(string)
	name := d.Get(constants.FieldCommonName).(string)
	obj, err := c.HarvesterClient.HarvesterhciV1beta1().Addons(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	if diags := updateAddon(ctx, c, d, namespace, obj); diags.HasError() {
		return diags
	}
	if err := waitForAddonApplied(ctx, c, d, namespace, name, schema.TimeoutCreate); err != nil {
		return diag.FromErr(err)
	}
	return resourceAddonRead(ctx, d, meta)
}

func resourceAddonRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterClient.HarvesterhciV1beta1().Addons(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceAddonImport(d, obj))
}

func resourceAddonUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterClient.HarvesterhciV1beta1().Addons(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if diags := updateAddon(ctx, c, d, namespace, obj); diags.HasError() {
		return diags
	}
	if err := waitForAddonApplied(ctx, c, d, namespace, name, schema.TimeoutUpdate); err != nil {
		return diag.FromErr(err)
	}
	return resourceAddonRead(ctx, d, meta)
}

// The addon cannot be deleted. It can only be disabled.
func resourceAddonDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterClient.HarvesterhciV1beta1().Addons(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if obj.Spec.Enabled {
		// Only flip the enabled switch: adopted addons may carry values that
		// predate Terraform management, wiping them on destroy would be
		// destructive.
		objCopy := obj.DeepCopy()
		objCopy.Spec.Enabled = false
		if _, err := c.HarvesterClient.HarvesterhciV1beta1().Addons(namespace).Update(ctx, objCopy, metav1.UpdateOptions{}); err != nil {
			return diag.FromErr(err)
		}
		if err := waitForAddonDisabled(ctx, c, d, namespace, name, schema.TimeoutDelete); err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId("")
	return nil
}

func updateAddon(ctx context.Context, c *client.Client, d *schema.ResourceData, namespace string, oldAddon *harvsterv1.Addon) diag.Diagnostics {
	toUpdate, err := util.ResourceConstruct(ctx, d, Updater(oldAddon.DeepCopy()))
	if err != nil {
		return diag.FromErr(err)
	}
	newAddon := toUpdate.(*harvsterv1.Addon)
	newAddon, err = c.HarvesterClient.HarvesterhciV1beta1().Addons(namespace).Update(ctx, newAddon, metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceAddonImport(d, newAddon))
}

func resourceAddonImport(d *schema.ResourceData, obj *harvsterv1.Addon) error {
	stateGetter, err := importer.ResourceAddonStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}
