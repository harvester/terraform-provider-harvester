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
	"time"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

	// The upload action is served by the Harvester Steve API server (harvester-system/harvester:8443),
	// NOT by the standard Kubernetes API. We route through the K8s API service proxy to reach it.
	uploadURL := fmt.Sprintf("%s/api/v1/namespaces/harvester-system/services/https:harvester:8443/proxy/v1/harvesterhci.io.virtualmachineimages/%s/%s?action=upload&size=%d",
		c.RestConfig.Host, namespace, name, stat.Size())

	transport, err := rest.TransportFor(c.RestConfig)
	if err != nil {
		return fmt.Errorf("failed to create transport: %w", err)
	}
	httpClient := &http.Client{Transport: transport}

	if err := waitForUploadAction(ctx, httpClient, uploadURL); err != nil {
		return err
	}
	return doUpload(ctx, httpClient, uploadURL, filePath)
}

// waitForUploadAction polls the upload endpoint until the action becomes available.
// The Harvester controller must initialize the image before the Steve API
// exposes the upload action (condition Initialized=False required).
func waitForUploadAction(ctx context.Context, httpClient *http.Client, uploadURL string) error {
	var lastErr error
	for i := 0; i < 30; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(2 * time.Second):
			}
		}

		probeReq, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, http.NoBody)
		if err != nil {
			return fmt.Errorf("failed to create probe request: %w", err)
		}
		probeResp, err := httpClient.Do(probeReq)
		if err != nil {
			return fmt.Errorf("probe request failed: %w", err)
		}
		probeResp.Body.Close()

		if probeResp.StatusCode == http.StatusForbidden || probeResp.StatusCode == http.StatusNotFound {
			lastErr = fmt.Errorf("upload action not yet available (HTTP %d)", probeResp.StatusCode)
			continue
		}
		return nil
	}
	return fmt.Errorf("upload action not available after waiting: %w", lastErr)
}

// doUpload streams the file to the Harvester upload endpoint as multipart/form-data.
func doUpload(ctx context.Context, httpClient *http.Client, uploadURL, filePath string) error {
	file, err := os.Open(filePath)
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
			file.Close()
			return
		}
		_, err = io.Copy(part, file)
		file.Close()
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		pw.CloseWithError(writer.Close())
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, pr)
	if err != nil {
		pr.Close()
		return fmt.Errorf("failed to create upload request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusBadRequest {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("upload failed (HTTP %d): %s", resp.StatusCode, string(body))
}
