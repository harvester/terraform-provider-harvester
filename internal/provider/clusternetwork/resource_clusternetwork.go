package clusternetwork

import (
	"context"

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

func ResourceClusterNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterNetworkCreate,
		ReadContext:   resourceClusterNetworkRead,
		DeleteContext: resourceClusterNetworkDelete,
		UpdateContext: resourceClusterNetworkUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: Schema(),
	}
}

func resourceClusterNetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	name := d.Get(constants.FieldCommonName).(string)
	toCreate, err := util.ResourceConstruct(d, Creator("", name))
	if err != nil {
		return diag.FromErr(err)
	}
	clusterNetwork, err := c.HarvesterNetworkClient.NetworkV1beta1().ClusterNetworks().Create(ctx, toCreate.(*harvsternetworkv1.ClusterNetwork), metav1.CreateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceClusterNetworkImport(d, clusterNetwork)
}

func resourceClusterNetworkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	clusterNetwork, err := c.HarvesterNetworkClient.NetworkV1beta1().ClusterNetworks().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	toUpdate, err := util.ResourceConstruct(d, Updater(clusterNetwork))
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = c.HarvesterNetworkClient.NetworkV1beta1().ClusterNetworks().Update(ctx, toUpdate.(*harvsternetworkv1.ClusterNetwork), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceClusterNetworkRead(ctx, d, meta)
}

func resourceClusterNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	clusterNetwork, err := c.HarvesterNetworkClient.NetworkV1beta1().ClusterNetworks().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return resourceClusterNetworkImport(d, clusterNetwork)
}

func resourceClusterNetworkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.HarvesterNetworkClient.NetworkV1beta1().ClusterNetworks().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func resourceClusterNetworkImport(d *schema.ResourceData, obj *harvsternetworkv1.ClusterNetwork) diag.Diagnostics {
	stateGetter, err := importer.ResourceClusterNetworkStateGetter(obj)
	if err != nil {
		return nil
	}
	return diag.FromErr(util.ResourceStatesSet(d, stateGetter))
}
