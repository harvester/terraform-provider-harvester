package image

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

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
		images, err := c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return diag.FromErr(err)
		}
		var matchIndices []int
		for i, image := range images.Items {
			if image.Spec.DisplayName == displayName {
				matchIndices = append(matchIndices, i)
			}
		}
		if len(matchIndices) == 0 {
			return diag.FromErr(fmt.Errorf("no image with display_name %q found in namespace %q", displayName, namespace))
		}
		if len(matchIndices) > 1 {
			names := make([]string, len(matchIndices))
			for i, idx := range matchIndices {
				names[i] = images.Items[idx].Name
			}
			return diag.FromErr(fmt.Errorf("display_name %q matches %d images in namespace %q: %v — use 'name' to select a specific image",
				displayName, len(matchIndices), namespace, names))
		}
		return diag.FromErr(resourceImageImport(d, &images.Items[matchIndices[0]]))
	}

	return diag.FromErr(fmt.Errorf("must specify image %s or %s", constants.FieldCommonName, constants.FieldImageDisplayName))
}
