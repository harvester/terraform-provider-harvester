package ippool

import (
	"context"
	"time"

	loadbalancerv1 "github.com/harvester/harvester-load-balancer/pkg/apis/loadbalancer.harvesterhci.io/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func ResourceIPPool() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIPPoolCreate,
		ReadContext:   resourceIPPoolRead,
		UpdateContext: resourceIPPoolUpdate,
		DeleteContext: resourceIPPoolDelete,
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

func resourceIPPoolCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	name := data.Get(constants.FieldCommonName).(string)
	toCreate, err := util.ResourceConstruct(data, Creator(name))
	if err != nil {
		return diag.FromErr(err)
	}
	ippool, err := c.HarvesterLoadbalancerClient.
		LoadbalancerV1beta1().
		IPPools().
		Create(ctx, toCreate.(*loadbalancerv1.IPPool), metav1.CreateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId(helper.BuildID("", name))
	return diag.FromErr(resourceIPPoolImport(data, ippool))
}

func resourceIPPoolRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	_, name, err := helper.IDParts(data.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ippool, err := c.HarvesterLoadbalancerClient.
		LoadbalancerV1beta1().
		IPPools().
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(resourceIPPoolImport(data, ippool))
}

func resourceIPPoolUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	_, name, err := helper.IDParts(data.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	obj, err := c.HarvesterLoadbalancerClient.
		LoadbalancerV1beta1().
		IPPools().
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	toUpdate, err := util.ResourceConstruct(data, Updater(obj))
	if err != nil {
		return diag.FromErr(err)
	}

	ippool, err := c.HarvesterLoadbalancerClient.
		LoadbalancerV1beta1().
		IPPools().
		Update(ctx, toUpdate.(*loadbalancerv1.IPPool), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceIPPoolImport(data, ippool))
}

func resourceIPPoolDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	_, name, err := helper.IDParts(data.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = c.HarvesterLoadbalancerClient.
		LoadbalancerV1beta1().
		IPPools().
		Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return diag.FromErr(nil)
}

func resourceIPPoolImport(data *schema.ResourceData, obj *loadbalancerv1.IPPool) error {
	stateGetter, err := importer.ResourceIPPoolStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(data, stateGetter)
}
