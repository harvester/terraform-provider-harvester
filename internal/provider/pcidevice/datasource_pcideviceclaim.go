package pcidevice

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func DataSourcePCIDeviceClaim() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePCIDeviceClaimRead,
		Schema:      dataSourcePCIDeviceClaimSchema(),
	}
}

func dataSourcePCIDeviceClaimSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		constants.FieldCommonName: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the PCIDeviceClaim.",
		},
		constants.FieldCommonLabels: {
			Type:        schema.TypeMap,
			Computed:    true,
			Description: "Labels of the PCIDeviceClaim.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		constants.FieldPCIDeviceClaimNodeName: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Node where the PCI device is located.",
		},
		constants.FieldPCIDeviceClaimAddress: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "PCI address of the claimed device.",
		},
	}
}

func dataSourcePCIDeviceClaimRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get(constants.FieldCommonName).(string)

	dc, err := dynamic.NewForConfig(c.RestConfig)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create dynamic client: %w", err))
	}

	claim, err := dc.Resource(PCIDeviceClaimGVR).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get PCIDeviceClaim %s: %w", name, err))
	}

	spec, _ := claim.Object["spec"].(map[string]interface{})

	d.SetId(name)
	states := map[string]interface{}{
		constants.FieldCommonName:             name,
		constants.FieldPCIDeviceClaimNodeName: getField(spec, "nodeName"),
		constants.FieldPCIDeviceClaimAddress:  getField(spec, "address"),
	}

	if labels := claim.GetLabels(); len(labels) > 0 {
		states[constants.FieldCommonLabels] = labels
	}

	for key, value := range states {
		if err := d.Set(key, value); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
