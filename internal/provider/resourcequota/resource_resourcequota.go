package resourcequota

import (
	"context"
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

func ResourceResourceQuota() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceResourceQuotaCreate,
		ReadContext:   resourceResourceQuotaRead,
		UpdateContext: resourceResourceQuotaUpdate,
		DeleteContext: resourceResourceQuotaDelete,
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

func resourceResourceQuotaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	namespace := d.Get(constants.FieldCommonNamespace).(string)
	name := d.Get(constants.FieldCommonName).(string)

	resourceQuota := &harvsterv1.ResourceQuota{
		ObjectMeta: util.NewObjectMeta(namespace, name),
		Spec: harvsterv1.ResourceQuotaSpec{
			SnapshotLimit: buildSnapshotLimit(d),
		},
	}

	setCommonFields(d, resourceQuota)

	obj, err := c.HarvesterClient.HarvesterhciV1beta1().ResourceQuotas(namespace).Create(ctx, resourceQuota, metav1.CreateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(resourceResourceQuotaImport(d, obj))
}

func resourceResourceQuotaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	obj, err := c.HarvesterClient.HarvesterhciV1beta1().ResourceQuotas(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	obj.Spec.SnapshotLimit = buildSnapshotLimit(d)
	setCommonFields(d, obj)

	_, err = c.HarvesterClient.HarvesterhciV1beta1().ResourceQuotas(namespace).Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceResourceQuotaRead(ctx, d, meta)
}

func resourceResourceQuotaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	obj, err := c.HarvesterClient.HarvesterhciV1beta1().ResourceQuotas(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return diag.FromErr(resourceResourceQuotaImport(d, obj))
}

func resourceResourceQuotaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err = c.HarvesterClient.HarvesterhciV1beta1().ResourceQuotas(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return diag.FromErr(err)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{constants.StateCommonActive},
		Target:     []string{constants.StateCommonRemoved},
		Refresh:    resourceResourceQuotaRefresh(ctx, d, meta),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      1 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceResourceQuotaImport(d *schema.ResourceData, obj *harvsterv1.ResourceQuota) error {
	stateGetter, err := importer.ResourceResourceQuotaStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}

func resourceResourceQuotaRefresh(ctx context.Context, d *schema.ResourceData, meta interface{}) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		c, err := meta.(*config.Config).K8sClient()
		if err != nil {
			return nil, constants.StateCommonError, err
		}
		namespace, name, err := helper.IDParts(d.Id())
		if err != nil {
			return nil, constants.StateCommonError, err
		}

		obj, err := c.HarvesterClient.HarvesterhciV1beta1().ResourceQuotas(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return obj, constants.StateCommonRemoved, nil
			}
			return obj, constants.StateCommonError, err
		}
		if err = resourceResourceQuotaImport(d, obj); err != nil {
			return obj, constants.StateCommonError, err
		}
		return obj, constants.StateCommonActive, nil
	}
}

func buildSnapshotLimit(d *schema.ResourceData) harvsterv1.SnapshotLimit {
	limit := harvsterv1.SnapshotLimit{}

	if v, ok := d.GetOk(constants.FieldResourceQuotaNamespaceTotalSnapshotSizeQuota); ok {
		limit.NamespaceTotalSnapshotSizeQuota = int64(v.(int))
	}

	if v, ok := d.GetOk(constants.FieldResourceQuotaVMTotalSnapshotSizeQuota); ok {
		vmQuotas := make(map[string]int64)
		for k, val := range v.(map[string]interface{}) {
			vmQuotas[k] = int64(val.(int))
		}
		limit.VMTotalSnapshotSizeQuota = vmQuotas
	}

	return limit
}

func setCommonFields(d *schema.ResourceData, obj *harvsterv1.ResourceQuota) {
	if v, ok := d.GetOk(constants.FieldCommonDescription); ok {
		if obj.Annotations == nil {
			obj.Annotations = map[string]string{}
		}
		obj.Annotations["field.cattle.io/description"] = v.(string)
	}

	if v, ok := d.GetOk(constants.FieldCommonTags); ok {
		for k, val := range v.(map[string]interface{}) {
			obj.Labels["tags.harvesterhci.io/"+k] = val.(string)
		}
	}

	if v, ok := d.GetOk(constants.FieldCommonLabels); ok {
		for k, val := range v.(map[string]interface{}) {
			obj.Labels[k] = val.(string)
		}
	}
}
