package vlanconfig

import (
	"context"
	"fmt"

	harvsternetworkv1 "github.com/harvester/harvester-network-controller/pkg/apis/network.harvesterhci.io/v1beta1"
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

const (
	vlanConfigDeleteTimeout = 300
)

func ResourceVLANConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVLANConfigCreate,
		ReadContext:   resourceVLANConfigRead,
		DeleteContext: resourceVLANConfigDelete,
		UpdateContext: resourceVLANConfigUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: Schema(),
	}
}

func resourceVLANConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	name := d.Get(constants.FieldCommonName).(string)
	toCreate, err := util.ResourceConstruct(d, Creator(name))
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterNetworkClient.NetworkV1beta1().VlanConfigs().Create(ctx, toCreate.(*harvsternetworkv1.VlanConfig), metav1.CreateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(helper.BuildID("", name))
	return diag.FromErr(resourceVLANConfigImport(d, obj))
}

func resourceVLANConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterNetworkClient.NetworkV1beta1().VlanConfigs().Get(ctx, name, metav1.GetOptions{})
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
	_, err = c.HarvesterNetworkClient.NetworkV1beta1().VlanConfigs().Update(ctx, toUpdate.(*harvsternetworkv1.VlanConfig), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceVLANConfigRead(ctx, d, meta)
}

func resourceVLANConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterNetworkClient.NetworkV1beta1().VlanConfigs().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceVLANConfigImport(d, obj))
}

func resourceVLANConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if err = c.HarvesterNetworkClient.NetworkV1beta1().VlanConfigs().Delete(ctx, name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return diag.FromErr(err)
	}
	events, err := c.HarvesterNetworkClient.NetworkV1beta1().VlanConfigs().Watch(ctx, util.WatchOptions(name, int64(vlanConfigDeleteTimeout)))
	if err != nil {
		return diag.FromErr(err)
	}
	if !util.HasDeleted(events) {
		return diag.FromErr(fmt.Errorf("timeout waiting for vlanconfig %s to be deleted", d.Id()))
	}
	d.SetId("")
	return nil
}

func resourceVLANConfigImport(d *schema.ResourceData, obj *harvsternetworkv1.VlanConfig) error {
	stateGetter, err := importer.ResourceVLANConfigStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}
