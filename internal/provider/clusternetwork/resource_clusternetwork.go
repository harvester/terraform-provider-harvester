package clusternetwork

import (
	"context"
	"fmt"
	"time"

	harvsternetworkv1 "github.com/harvester/harvester-network-controller/pkg/apis/network.harvesterhci.io/v1beta1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/util"
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
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(2 * time.Minute),
			Read:    schema.DefaultTimeout(2 * time.Minute),
			Update:  schema.DefaultTimeout(2 * time.Minute),
			Delete:  schema.DefaultTimeout(2 * time.Minute),
			Default: schema.DefaultTimeout(2 * time.Minute),
		},
	}
}

func resourceClusterNetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	if name == constants.ManagementClusterNetworkName {
		return diag.FromErr(fmt.Errorf("can not create the existing %s clusternetwork, to avoid this error and continue with the plan, use `terraform import harvester_clusternetwork.%s %s` to import it first", name, name, name))
	}
	toCreate, err := util.ResourceConstruct(d, Creator(name))
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterNetworkClient.NetworkV1beta1().ClusterNetworks().Create(ctx, toCreate.(*harvsternetworkv1.ClusterNetwork), metav1.CreateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(helper.BuildID("", name))
	return diag.FromErr(resourceClusterNetworkImport(d, obj))
}

func resourceClusterNetworkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterNetworkClient.NetworkV1beta1().ClusterNetworks().Get(ctx, name, metav1.GetOptions{})
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
	_, err = c.HarvesterNetworkClient.NetworkV1beta1().ClusterNetworks().Update(ctx, toUpdate.(*harvsternetworkv1.ClusterNetwork), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceClusterNetworkRead(ctx, d, meta)
}

func resourceClusterNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterNetworkClient.NetworkV1beta1().ClusterNetworks().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceClusterNetworkImport(d, obj))
}

func resourceClusterNetworkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if name == constants.ManagementClusterNetworkName {
		return diag.FromErr(fmt.Errorf("clusternetwork %s should not be destroyed, to avoid this error and continue with the plan, either move clusternetwork %s to another module or reduce the scope of the plan using the -target flag", name, name))
	}

	err = c.HarvesterNetworkClient.NetworkV1beta1().ClusterNetworks().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func resourceClusterNetworkImport(d *schema.ResourceData, obj *harvsternetworkv1.ClusterNetwork) error {
	stateGetter, err := importer.ResourceClusterNetworkStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}
