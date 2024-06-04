package network

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nadv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func ResourceNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkCreate,
		ReadContext:   resourceNetworkRead,
		DeleteContext: resourceNetworkDelete,
		UpdateContext: resourceNetworkUpdate,
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

func resourceNetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	namespace := d.Get(constants.FieldCommonNamespace).(string)
	name := d.Get(constants.FieldCommonName).(string)
	toCreate, err := util.ResourceConstruct(d, Creator(c, ctx, namespace, name))
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterClient.K8sCniCncfIoV1().NetworkAttachmentDefinitions(namespace).Create(ctx, toCreate.(*nadv1.NetworkAttachmentDefinition), metav1.CreateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceNetworkImport(d, obj))
}

func resourceNetworkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterClient.K8sCniCncfIoV1().NetworkAttachmentDefinitions(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	toUpdate, err := util.ResourceConstruct(d, Updater(c, ctx, obj))
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = c.HarvesterClient.K8sCniCncfIoV1().NetworkAttachmentDefinitions(namespace).Update(ctx, toUpdate.(*nadv1.NetworkAttachmentDefinition), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceNetworkRead(ctx, d, meta)
}

func resourceNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterClient.K8sCniCncfIoV1().NetworkAttachmentDefinitions(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return diag.FromErr(resourceNetworkImport(d, obj))
}

func resourceNetworkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if err = c.HarvesterClient.K8sCniCncfIoV1().NetworkAttachmentDefinitions(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return diag.FromErr(err)
	}

	ctxDeadline, _ := ctx.Deadline()
	events, err := c.HarvesterClient.
		K8sCniCncfIoV1().
		NetworkAttachmentDefinitions(namespace).
		Watch(ctx, util.WatchOptions(name, time.Until(ctxDeadline)))
	if err != nil {
		return diag.FromErr(err)
	}
	if !util.HasDeleted(events) {
		return diag.FromErr(fmt.Errorf("timeout waiting for network %s to be deleted", d.Id()))
	}

	d.SetId("")
	return nil
}

func resourceNetworkImport(d *schema.ResourceData, obj *nadv1.NetworkAttachmentDefinition) error {
	stateGetter, err := importer.ResourceNetworkStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}
