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

func DataSourcePCIDevice() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePCIDeviceRead,
		Schema:      dataSourcePCIDeviceSchema(),
	}
}

func dataSourcePCIDeviceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		constants.FieldCommonName: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the PCIDevice (format: '{nodeName}-{addressWithoutColonsAndDots}').",
		},
		constants.FieldPCIDeviceAddress: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "PCI address (e.g., '0000:01:00.0').",
		},
		constants.FieldPCIDeviceNodeName: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Node where the PCI device is located.",
		},
		constants.FieldPCIDeviceVendorID: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "PCI vendor ID (e.g., '8086' for Intel).",
		},
		constants.FieldPCIDeviceDeviceID: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "PCI device ID.",
		},
		constants.FieldPCIDeviceClassID: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "PCI class ID (e.g., '0300' for VGA).",
		},
		constants.FieldPCIDeviceDeviceDescription: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Human-readable device description.",
		},
		constants.FieldPCIDeviceIOMMUGroup: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "IOMMU group of the device.",
		},
		constants.FieldPCIDeviceKernelDriver: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Kernel driver currently in use.",
		},
		constants.FieldPCIDeviceResourceName: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Kubernetes device plugin resource name.",
		},
	}
}

func dataSourcePCIDeviceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get(constants.FieldCommonName).(string)

	dc, err := dynamic.NewForConfig(c.RestConfig)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create dynamic client: %w", err))
	}

	device, err := dc.Resource(PCIDeviceGVR).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get PCIDevice %s: %w", name, err))
	}

	// PCIDevice has all info in status, not spec
	status, _ := device.Object["status"].(map[string]interface{})

	d.SetId(name)
	states := map[string]interface{}{
		constants.FieldPCIDeviceAddress:           getField(status, "address"),
		constants.FieldPCIDeviceNodeName:          getField(status, "nodeName"),
		constants.FieldPCIDeviceVendorID:          getField(status, "vendorId"),
		constants.FieldPCIDeviceDeviceID:          getField(status, "deviceId"),
		constants.FieldPCIDeviceClassID:           getField(status, "classId"),
		constants.FieldPCIDeviceDeviceDescription: getField(status, "description"),
		constants.FieldPCIDeviceIOMMUGroup:        getField(status, "iommuGroup"),
		constants.FieldPCIDeviceKernelDriver:      getField(status, "kernelDriverInUse"),
		constants.FieldPCIDeviceResourceName:      getField(status, "resourceName"),
	}
	for key, value := range states {
		if err := d.Set(key, value); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
