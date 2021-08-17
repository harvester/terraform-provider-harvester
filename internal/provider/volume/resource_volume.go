package volume

import (
	"context"

	corev1 "k8s.io/api/core/v1"

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

func ResourceVolume() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVolumeCreate,
		ReadContext:   resourceVolumeRead,
		DeleteContext: resourceVolumeDelete,
		UpdateContext: resourceVolumeUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: Schema(),
	}
}

func resourceVolumeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	namespace := d.Get(constants.FieldCommonNamespace).(string)
	name := d.Get(constants.FieldCommonName).(string)
	toCreate, err := util.ResourceConstruct(d, Creator(namespace, name))
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, toCreate.(*corev1.PersistentVolumeClaim), metav1.CreateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceVolumeImport(d, obj)
}

func resourceVolumeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	obj, err := c.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
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
	_, err = c.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Update(ctx, toUpdate.(*corev1.PersistentVolumeClaim), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceVolumeRead(ctx, d, meta)
}

func resourceVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return resourceVolumeImport(d, obj)
}

func resourceVolumeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func resourceVolumeImport(d *schema.ResourceData, obj *corev1.PersistentVolumeClaim) diag.Diagnostics {
	stateGetter, err := importer.ResourceVolumeStateGetter(obj)
	if err != nil {
		return nil
	}
	return diag.FromErr(util.ResourceStatesSet(d, stateGetter))
}
