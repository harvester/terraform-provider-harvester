package image

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func DataSourceImage() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceImageRead,
		Schema:      DataSourceSchema(),
	}
}

func dataSourceImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
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
		for i, image := range images.Items {
			if image.Spec.DisplayName == displayName {
				return diag.FromErr(resourceImageImport(d, &images.Items[i]))
			}
		}
		return diag.FromErr(fmt.Errorf("can not find image %s in namespace %s", displayName, namespace))
	}

	return diag.FromErr(fmt.Errorf("must specify image %s or %s", constants.FieldCommonName, constants.FieldImageDisplayName))
}
