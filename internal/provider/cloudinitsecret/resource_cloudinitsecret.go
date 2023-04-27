package cloudinitsecret

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func ResourceCloudInitSecret() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudInitSecretCreate,
		ReadContext:   resourceCloudInitSecretRead,
		DeleteContext: resourceCloudInitSecretDelete,
		UpdateContext: resourceCloudInitSecretUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: Schema(),
	}
}

func resourceCloudInitSecretCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	namespace := d.Get(constants.FieldCommonNamespace).(string)
	name := d.Get(constants.FieldCommonName).(string)
	toCreate, err := util.ResourceConstruct(d, Creator(namespace, name))
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeClient.CoreV1().Secrets(namespace).Create(ctx, toCreate.(*corev1.Secret), metav1.CreateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(resourceCloudInitSecretImport(d, obj))
}

func resourceCloudInitSecretUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeClient.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
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
	_, err = c.KubeClient.CoreV1().Secrets(namespace).Update(ctx, toUpdate.(*corev1.Secret), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceCloudInitSecretRead(ctx, d, meta)
}

func resourceCloudInitSecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.KubeClient.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceCloudInitSecretImport(d, obj))
}

func resourceCloudInitSecretDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.KubeClient.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func resourceCloudInitSecretImport(d *schema.ResourceData, obj *corev1.Secret) error {
	stateGetter, err := importer.ResourceCloudInitSecretStateGetter(obj)
	if err != nil {
		return err
	}

	if d.Get(constants.FieldCloudInitSecretUserData) != "" && d.Get(constants.FieldCloudInitSecretUserDataBase64) == "" {
		userdata, err := base64.StdEncoding.DecodeString(stateGetter.States[constants.FieldCloudInitSecretUserDataBase64].(string))
		if err != nil {
			return fmt.Errorf("can't decode userdata: %v", err)
		}
		stateGetter.States[constants.FieldCloudInitSecretUserData] = string(userdata)
		stateGetter.States[constants.FieldCloudInitSecretUserDataBase64] = ""
	}
	if d.Get(constants.FieldCloudInitSecretNetworkData) != "" && d.Get(constants.FieldCloudInitSecretNetworkDataBase64) == "" {
		networkdata, err := base64.StdEncoding.DecodeString(stateGetter.States[constants.FieldCloudInitSecretNetworkDataBase64].(string))
		if err != nil {
			return fmt.Errorf("can't decode userdata: %v", err)
		}
		stateGetter.States[constants.FieldCloudInitSecretUserData] = string(networkdata)
		stateGetter.States[constants.FieldCloudInitSecretUserDataBase64] = ""
	}

	return util.ResourceStatesSet(d, stateGetter)
}
