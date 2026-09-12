package blockdevice

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

var blockDeviceGVR = k8sschema.GroupVersionResource{
	Group:    "harvesterhci.io",
	Version:  "v1beta1",
	Resource: "blockdevices",
}

func ResourceBlockDevice() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBlockDeviceCreate,
		ReadContext:   resourceBlockDeviceRead,
		UpdateContext: resourceBlockDeviceUpdate,
		DeleteContext: resourceBlockDeviceDelete,
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

// resourceBlockDeviceCreate adopts an existing block device by applying spec updates.
func resourceBlockDeviceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	namespace := d.Get(constants.FieldCommonNamespace).(string)
	name := d.Get(constants.FieldCommonName).(string)

	obj, err := c.DynamicClient.Resource(blockDeviceGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(fmt.Errorf("block device %s/%s not found, it must already exist on the node: %w", namespace, name, err))
	}

	applyBlockDeviceSpec(d, obj)

	obj, err = c.DynamicClient.Resource(blockDeviceGVR).Namespace(namespace).Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(resourceBlockDeviceImport(d, obj))
}

func resourceBlockDeviceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	obj, err := c.DynamicClient.Resource(blockDeviceGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return diag.FromErr(resourceBlockDeviceImport(d, obj))
}

func resourceBlockDeviceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	obj, err := c.DynamicClient.Resource(blockDeviceGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	applyBlockDeviceSpec(d, obj)

	obj, err = c.DynamicClient.Resource(blockDeviceGVR).Namespace(namespace).Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(resourceBlockDeviceImport(d, obj))
}

// resourceBlockDeviceDelete deprovisions the device but does NOT delete the K8s object.
func resourceBlockDeviceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	obj, err := c.DynamicClient.Resource(blockDeviceGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// Unprovision: set provision=false and clear provisioner config
	_ = unstructured.SetNestedField(obj.Object, false, "spec", "provision")
	unstructured.RemoveNestedField(obj.Object, "spec", "provisioner")
	unstructured.RemoveNestedField(obj.Object, "spec", "tags")

	_, err = c.DynamicClient.Resource(blockDeviceGVR).Namespace(namespace).Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceBlockDeviceImport(d *schema.ResourceData, obj *unstructured.Unstructured) error {
	stateGetter, err := importer.ResourceBlockDeviceStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}

// applyBlockDeviceSpec applies user-specified spec fields to the unstructured object.
func applyBlockDeviceSpec(d *schema.ResourceData, obj *unstructured.Unstructured) {
	provision := d.Get(constants.FieldBlockDeviceProvision).(bool)
	_ = unstructured.SetNestedField(obj.Object, provision, "spec", "provision")

	forceFormatted := d.Get(constants.FieldBlockDeviceForceFormatted).(bool)
	_ = unstructured.SetNestedField(obj.Object, forceFormatted, "spec", "fileSystem", "forceFormatted")

	// Device tags (spec.tags)
	if v, ok := d.GetOk(constants.FieldBlockDeviceDeviceTags); ok {
		rawTags := v.([]interface{})
		tags := make([]interface{}, len(rawTags))
		copy(tags, rawTags)
		_ = unstructured.SetNestedSlice(obj.Object, tags, "spec", "tags")
	} else {
		unstructured.RemoveNestedField(obj.Object, "spec", "tags")
	}

	// Provisioner block
	if v, ok := d.GetOk(constants.FieldBlockDeviceProvisioner); ok {
		provList := v.([]interface{})
		if len(provList) > 0 && provList[0] != nil {
			provMap := provList[0].(map[string]interface{})
			applyProvisioner(provMap, obj)
		} else {
			unstructured.RemoveNestedField(obj.Object, "spec", "provisioner")
		}
	} else {
		unstructured.RemoveNestedField(obj.Object, "spec", "provisioner")
	}

	// Apply description annotation
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	if desc, ok := d.GetOk(constants.FieldCommonDescription); ok {
		annotations["field.cattle.io/description"] = desc.(string)
	} else {
		delete(annotations, "field.cattle.io/description")
	}
	obj.SetAnnotations(annotations)

	// Apply user tags (harvesterhci.io tags on labels)
	labels := obj.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	if tagsRaw, ok := d.GetOk(constants.FieldCommonTags); ok {
		tags := tagsRaw.(map[string]interface{})
		for k, v := range tags {
			labels["tags.harvesterhci.io/"+k] = v.(string)
		}
	}
	obj.SetLabels(labels)
}

func applyProvisioner(provMap map[string]interface{}, obj *unstructured.Unstructured) {
	if lhList, ok := provMap[constants.FieldBlockDeviceProvisionerLonghorn]; ok {
		lhItems := lhList.([]interface{})
		if len(lhItems) > 0 && lhItems[0] != nil {
			lh := lhItems[0].(map[string]interface{})
			provisioner := map[string]interface{}{
				"longhorn": map[string]interface{}{
					"engineVersion": lh[constants.FieldBlockDeviceProvisionerLonghornEV],
					"diskDriver":    lh[constants.FieldBlockDeviceProvisionerLonghornDD],
				},
			}
			_ = unstructured.SetNestedField(obj.Object, provisioner, "spec", "provisioner")
			return
		}
	}

	if lvmList, ok := provMap[constants.FieldBlockDeviceProvisionerLVM]; ok {
		lvmItems := lvmList.([]interface{})
		if len(lvmItems) > 0 && lvmItems[0] != nil {
			lvm := lvmItems[0].(map[string]interface{})
			provisionerMap := map[string]interface{}{
				"lvm": map[string]interface{}{
					"vgName": lvm[constants.FieldBlockDeviceProvisionerLVMVGName],
				},
			}
			if params, ok := lvm[constants.FieldBlockDeviceProvisionerLVMParameters]; ok {
				paramsList := params.([]interface{})
				if len(paramsList) > 0 {
					provisionerMap["lvm"].(map[string]interface{})["parameters"] = paramsList
				}
			}
			_ = unstructured.SetNestedField(obj.Object, provisionerMap, "spec", "provisioner")
			return
		}
	}

	unstructured.RemoveNestedField(obj.Object, "spec", "provisioner")
}
