package image

import (
	"context"
	"fmt"
	"strings"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	harvsterutil "github.com/harvester/harvester/pkg/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// isValidLabelValue returns true if the string is a valid Kubernetes label value.
func isValidLabelValue(v string) bool {
	return len(validation.IsValidLabelValue(v)) == 0
}

// imagesMatchingDisplayName returns the images whose spec.displayName equals
// displayName. The comparison always uses the spec: the imageDisplayName label
// only mirrors it for display names that are valid label values, and only once
// the image import succeeded.
func imagesMatchingDisplayName(items []harvsterv1.VirtualMachineImage, displayName string) []*harvsterv1.VirtualMachineImage {
	var matches []*harvsterv1.VirtualMachineImage
	for i := range items {
		if items[i].Spec.DisplayName == displayName {
			matches = append(matches, &items[i])
		}
	}
	return matches
}

func lookupImageByDisplayName(ctx context.Context, c *client.Client, namespace, displayName string) (*harvsterv1.VirtualMachineImage, error) {
	imageClient := c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(namespace)

	var matches []*harvsterv1.VirtualMachineImage
	// Fast path: Harvester mirrors display names that are valid label values
	// into the imageDisplayName label, so the API server can filter for us.
	if isValidLabelValue(displayName) {
		images, err := imageClient.List(ctx, metav1.ListOptions{
			LabelSelector: harvsterutil.LabelImageDisplayName + "=" + displayName,
		})
		if err != nil {
			return nil, err
		}
		matches = imagesMatchingDisplayName(images.Items, displayName)
	}
	// The label is only set once the image import succeeded (and never for
	// display names that are not valid label values), so an empty fast path is
	// not conclusive: fall back to a full list before reporting anything.
	if len(matches) == 0 {
		images, err := imageClient.List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		matches = imagesMatchingDisplayName(images.Items, displayName)
	}

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("cannot find image with display_name %s in namespace %s", displayName, namespace)
	case 1:
		return matches[0], nil
	default:
		names := make([]string, len(matches))
		for i, image := range matches {
			names[i] = image.Name
		}
		// Harvester only enforces display name uniqueness through the
		// imageDisplayName label, so duplicates can exist for display names
		// that are not valid label values (e.g. containing spaces).
		return nil, fmt.Errorf("display_name %s matches %d images in namespace %s (%s), use name to select a specific image",
			displayName, len(matches), namespace, strings.Join(names, ", "))
	}
}

func DataSourceImage() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceImageRead,
		Schema:      DataSourceSchema(),
	}
}

func dataSourceImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace := d.Get(constants.FieldCommonNamespace).(string)
	name := d.Get(constants.FieldCommonName).(string)
	displayName := d.Get(constants.FieldImageDisplayName).(string)

	if name != "" {
		image, err := c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.FromErr(resourceImageImport(d, image))
	}

	if displayName != "" {
		image, err := lookupImageByDisplayName(ctx, c, namespace, displayName)
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.FromErr(resourceImageImport(d, image))
	}

	return diag.FromErr(fmt.Errorf("must specify image %s or %s", constants.FieldCommonName, constants.FieldImageDisplayName))
}
