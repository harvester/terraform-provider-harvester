package sriovdevice

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

func DataSourceSRIOVNetworkDevice() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSRIOVNetworkDeviceRead,
		Schema:      DataSourceSchema(),
	}
}

func dataSourceSRIOVNetworkDeviceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get(constants.FieldCommonName).(string)
	pcidevice, err := c.HarvesterDeviceClient.DevicesV1beta1().SRIOVNetworkDevices().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(resourceSRIOVNetworkDeviceImport(d, pcidevice))
}

func resourceSRIOVNetworkDeviceImport(d *schema.ResourceData, obj *devicesv1.SRIOVNetworkDevice) error {
	stateGetter, err := importer.ResourceSRIOVNetworkDeviceStateGetter(obj)
	if err != nil {
		return err
	}
	return util.ResourceStatesSet(d, stateGetter)
}
