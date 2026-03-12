package pcidevice

import (
	"context"

	devicesv1 "github.com/harvester/pcidevices/pkg/apis/devices.harvesterhci.io/v1beta1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func DataSourcePCIDevice() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePCIDeviceRead,
		Schema:      DataSourceSchema(),
	}
}

func dataSourcePCIDeviceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	pcidevice, err := c.HarvesterDeviceClient.DevicesV1beta1().PCIDevices().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourcePCIDeviceImport(d, pcidevice))
}

func resourcePCIDeviceImport(d *schema.ResourceData, obj *devicesv1.PCIDevice) error {
	stateGetter, err := importer.ResourcePCIDeviceStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}
