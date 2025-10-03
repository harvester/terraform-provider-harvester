package loadbalancer

import (
	"context"
	"errors"
	"time"

	loadbalancerv1 "github.com/harvester/harvester-load-balancer/pkg/apis/loadbalancer.harvesterhci.io/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func ResourceLoadBalancer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLoadBalancerCreate,
		ReadContext:   resourceLoadBalancerRead,
		UpdateContext: resourceLoadBalancerUpdate,
		DeleteContext: resourceLoadBalancerDelete,
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

func resourceLoadBalancerCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace := data.Get(constants.FieldCommonNamespace).(string)
	name := data.Get(constants.FieldCommonName).(string)
	toCreate, err := util.ResourceConstruct(data, Creator(namespace, name))
	if err != nil {
		return diag.FromErr(err)
	}

	loadbalancer, err := c.HarvesterLoadbalancerClient.
		LoadbalancerV1beta1().
		LoadBalancers(namespace).
		Create(ctx, toCreate.(*loadbalancerv1.LoadBalancer), metav1.CreateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(helper.BuildID(namespace, name))

	if resourceLoadBalancerSetIPAddress(ctx, data, meta) != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(resourceLoadBalancerImport(data, loadbalancer))
}

func resourceLoadBalancerSetIPAddress(ctx context.Context, data *schema.ResourceData, meta interface{}) error {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return err
	}

	retryInterval := 3 * time.Second
	retryTimeout := 10

	namespace, name, err := helper.IDParts(data.Id())
	if err != nil {
		return err
	}

	for i := 0; i < retryTimeout; i++ {
		loadbalancer, err := c.HarvesterLoadbalancerClient.
			LoadbalancerV1beta1().
			LoadBalancers(namespace).
			Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		// If LB or Status is missing, retry
		if loadbalancer == nil || &loadbalancer.Status == (&loadbalancerv1.LoadBalancerStatus{}) {
			time.Sleep(retryInterval)
			continue
		}

		if loadbalancer.Status.Address != "" {
			if err := data.Set(constants.FieldLoadBalancerIPAddress, loadbalancer.Status.Address); err != nil {
				return err
			}
			return nil
		}
		time.Sleep(retryInterval)
	}

	return errors.New("no address was populated for the loadbalancer")
}

func resourceLoadBalancerRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace, name, err := helper.IDParts(data.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	loadbalancer, err := c.HarvesterLoadbalancerClient.
		LoadbalancerV1beta1().
		LoadBalancers(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	if resourceLoadBalancerSetIPAddress(ctx, data, meta) != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(resourceLoadBalancerImport(data, loadbalancer))
}

func resourceLoadBalancerUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace, name, err := helper.IDParts(data.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	obj, err := c.HarvesterLoadbalancerClient.
		LoadbalancerV1beta1().
		LoadBalancers(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	toUpdate, err := util.ResourceConstruct(data, Updater(obj))
	if err != nil {
		return diag.FromErr(err)
	}

	loadbalancer, err := c.HarvesterLoadbalancerClient.
		LoadbalancerV1beta1().
		LoadBalancers(namespace).
		Update(ctx, toUpdate.(*loadbalancerv1.LoadBalancer), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	if resourceLoadBalancerSetIPAddress(ctx, data, meta) != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(resourceLoadBalancerImport(data, loadbalancer))
}

func resourceLoadBalancerDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace, name, err := helper.IDParts(data.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = c.HarvesterLoadbalancerClient.
		LoadbalancerV1beta1().
		LoadBalancers(namespace).
		Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return diag.FromErr(err)
	}
	return diag.FromErr(nil)
}

func resourceLoadBalancerImport(data *schema.ResourceData, obj *loadbalancerv1.LoadBalancer) error {
	stateGetter, err := importer.ResourceLoadBalancerStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(data, stateGetter)
}
