package image

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func ResourceImage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceImageCreate,
		ReadContext:   resourceImageRead,
		DeleteContext: resourceImageDelete,
		UpdateContext: resourceImageUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: Schema(),
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(5 * time.Minute),
			Read:    schema.DefaultTimeout(2 * time.Minute),
			Update:  schema.DefaultTimeout(5 * time.Minute),
			Delete:  schema.DefaultTimeout(2 * time.Minute),
			Default: schema.DefaultTimeout(2 * time.Minute),
		},
	}
}

func resourceImageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace := d.Get(constants.FieldCommonNamespace).(string)
	name := d.Get(constants.FieldCommonName).(string)
	toCreate, err := util.ResourceConstruct(d, Creator(namespace, name))
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(namespace).Create(ctx, toCreate.(*harvsterv1.VirtualMachineImage), metav1.CreateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(helper.BuildID(namespace, name))

	if d.Get(constants.FieldImageSourceType).(string) == string(harvsterv1.VirtualMachineImageSourceTypeUpload) {
		filePath := d.Get(constants.FieldImageFilePath).(string)
		if err := uploadImageFile(ctx, c, namespace, name, filePath); err != nil {
			_ = c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(namespace).Delete(ctx, name, metav1.DeleteOptions{})
			return diag.FromErr(fmt.Errorf("failed to upload image: %w", err))
		}
	}

	return diag.FromErr(resourceImageWaitForState(ctx, d, meta, schema.TimeoutCreate))
}

func resourceImageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(namespace).Get(ctx, name, metav1.GetOptions{})
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
	_, err = c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(namespace).Update(ctx, toUpdate.(*harvsterv1.VirtualMachineImage), metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceImageRead(ctx, d, meta)
}

func resourceImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	obj, err := c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceImageImport(d, obj))
}

func resourceImageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace, name, err := helper.IDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return diag.FromErr(err)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{constants.StateImageTerminating, constants.StateCommonActive},
		Target:     []string{constants.StateCommonRemoved},
		Refresh:    resourceImageRefresh(ctx, d, meta),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      1 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceImageImport(d *schema.ResourceData, obj *harvsterv1.VirtualMachineImage) error {
	stateGetter, err := importer.ResourceImageStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}

func resourceImageWaitForState(ctx context.Context, d *schema.ResourceData, meta interface{}, timeOutKey string) error {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{constants.StateImageInitializing, constants.StateImageDownloading, constants.StateImageUploading, constants.StateImageExporting},
		Target:     []string{constants.StateCommonActive},
		Refresh:    resourceImageRefresh(ctx, d, meta),
		Timeout:    d.Timeout(timeOutKey),
		Delay:      1 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func resourceImageRefresh(ctx context.Context, d *schema.ResourceData, meta interface{}) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		c, err := meta.(*config.Config).K8sClient()
		if err != nil {
			return nil, "", err
		}
		namespace := d.Get(constants.FieldCommonNamespace).(string)
		name := d.Get(constants.FieldCommonName).(string)
		obj, err := c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return obj, constants.StateCommonRemoved, nil
			}
			return obj, constants.StateCommonError, err
		}
		if err = resourceImageImport(d, obj); err != nil {
			return obj, constants.StateCommonError, err
		}
		state := d.Get(constants.FieldCommonState).(string)
		if state == constants.StateCommonFailed {
			message := d.Get(constants.FieldCommonMessage).(string)
			return obj, state, errors.New(message)
		}
		return obj, state, err
	}
}

func uploadImageFile(ctx context.Context, c *client.Client, namespace, name, filePath string) error {
	stat, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}

	// Wait for the Harvester controller to initialize the image before uploading.
	// The upload action is only available after Initialized=True on the VMI.
	if err := waitForImageInitialized(ctx, c, namespace, name); err != nil {
		return err
	}

	// The upload action is served by Harvester's Steve API, exposed directly at /v1/harvester/.
	uploadURL := fmt.Sprintf("%s/v1/harvester/harvesterhci.io.virtualmachineimages/%s/%s?action=upload&size=%d",
		c.RestConfig.Host, namespace, name, stat.Size())

	// Use rest.TransportFor to get a transport pre-configured with TLS, proxy, and auth from kubeconfig.
	transport, err := rest.TransportFor(c.RestConfig)
	if err != nil {
		return fmt.Errorf("failed to create transport: %w", err)
	}

	// Retry upload on transient errors: "already exists" (HTTP 400) can occur when a previously
	// deleted image's backing volume hasn't been fully cleaned up yet, and "timeout waiting"
	// (HTTP 500) can occur when the backing image data source pod needs more time to start.
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		err = doUpload(ctx, transport, uploadURL, filePath)
		if err == nil {
			return nil
		}
		if !strings.Contains(err.Error(), "already exists") && !strings.Contains(err.Error(), "timeout waiting") {
			return err
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("upload failed: %w", err)
		case <-ticker.C:
		}
	}
}

// waitForImageInitialized polls the VirtualMachineImage status until the Initialized
// condition is True. The Harvester controller must initialize the image (create the
// backing image and data source) before the Steve API exposes the upload action.
func waitForImageInitialized(ctx context.Context, c *client.Client, namespace, name string) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		obj, err := c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to check image status: %w", err)
		}
		for _, cond := range obj.Status.Conditions {
			if cond.Type == harvsterv1.ImageInitialized && string(cond.Status) == string(corev1.ConditionTrue) {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("image not initialized after waiting: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

// doUpload streams the file to the Harvester upload endpoint as multipart/form-data.
func doUpload(ctx context.Context, transport http.RoundTripper, uploadURL, filePath string) error {
	file, err := os.Open(filePath) //nolint:gosec // G304: filePath is user-provided via Terraform configuration
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}

	// Send only the base filename to avoid leaking the local filesystem path.
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	go func() {
		part, err := writer.CreateFormFile("chunk", filepath.Base(filePath))
		if err != nil {
			pw.CloseWithError(err)
			_ = file.Close()
			return
		}
		_, err = io.Copy(part, file)
		_ = file.Close()
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		pw.CloseWithError(writer.Close())
	}()

	resp, err := util.DoPostWithTransport(ctx, uploadURL, pr, writer.FormDataContentType(), transport)
	if err != nil {
		_ = pr.Close()
		return fmt.Errorf("upload request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusBadRequest {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("upload failed (HTTP %d): %s", resp.StatusCode, string(body))
}
