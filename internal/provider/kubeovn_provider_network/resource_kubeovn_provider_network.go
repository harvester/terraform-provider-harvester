package kubeovn_provider_network

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	kubeovnclient "github.com/kubeovn/kube-ovn/pkg/client/clientset/versioned/typed/kubeovn/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func ResourceKubeOVNProviderNetwork() *schema.Resource {
	return &schema.Resource{
		Description:   "A kube-ovn provider network bridges a host interface of every node into kube-ovn, for VLAN (underlay) subnets.",
		CreateContext: resourceKubeOVNProviderNetworkCreate,
		ReadContext:   resourceKubeOVNProviderNetworkRead,
		UpdateContext: resourceKubeOVNProviderNetworkUpdate,
		DeleteContext: resourceKubeOVNProviderNetworkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: Schema(),
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(5 * time.Minute),
			Read:    schema.DefaultTimeout(2 * time.Minute),
			Update:  schema.DefaultTimeout(2 * time.Minute),
			Delete:  schema.DefaultTimeout(2 * time.Minute),
			Default: schema.DefaultTimeout(2 * time.Minute),
		},
	}
}

func resourceKubeOVNProviderNetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	toCreate, err := util.ResourceConstruct(ctx, d, Creator(name))
	if err != nil {
		return diag.FromErr(err)
	}
	providerNetworks := c.KubeOVNClient.KubeovnV1().ProviderNetworks()
	if _, err = providerNetworks.Create(ctx, toCreate.(*kubeovnv1.ProviderNetwork), metav1.CreateOptions{}); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(helper.BuildID("", name))
	// kube-ovn sets status.ready once the bridge is up on every node that is
	// not excluded: a VLAN or a subnet on it is not usable before.
	stateConf := &retry.StateChangeConf{
		Pending:    []string{constants.StateCommonActive},
		Target:     []string{constants.StateCommonReady},
		Refresh:    providerNetworkRefresh(ctx, providerNetworks, name),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      1 * time.Second,
		MinTimeout: 2 * time.Second,
	}
	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("waiting for kube-ovn provider network %s to be ready: %v", name, err)
	}
	return resourceKubeOVNProviderNetworkRead(ctx, d, meta)
}

func providerNetworkRefresh(ctx context.Context, providerNetworks kubeovnclient.ProviderNetworkInterface, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		providerNetwork, err := providerNetworks.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return providerNetwork, constants.StateCommonRemoved, nil
			}
			return providerNetwork, constants.StateCommonError, err
		}
		return providerNetwork, importer.KubeOVNReadyState(providerNetwork.Status.Ready), nil
	}
}

func resourceKubeOVNProviderNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeOVNClient.KubeovnV1().ProviderNetworks().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceKubeOVNProviderNetworkImport(d, obj))
}

func resourceKubeOVNProviderNetworkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeOVNClient.KubeovnV1().ProviderNetworks().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	toUpdate, err := util.ResourceConstruct(ctx, d, Updater(obj))
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = c.KubeOVNClient.KubeovnV1().ProviderNetworks().Update(ctx, toUpdate.(*kubeovnv1.ProviderNetwork), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceKubeOVNProviderNetworkRead(ctx, d, meta)
}

func resourceKubeOVNProviderNetworkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	_, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	// kube-ovn puts no finalizer on a provider network: it is gone once the
	// delete call returns.
	err = c.KubeOVNClient.KubeovnV1().ProviderNetworks().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return diag.Errorf("deleting kube-ovn provider network %s: %v", name, err)
	}
	d.SetId("")
	return nil
}

func resourceKubeOVNProviderNetworkImport(d *schema.ResourceData, obj *kubeovnv1.ProviderNetwork) error {
	stateGetter, err := importer.ResourceKubeOVNProviderNetworkStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}
