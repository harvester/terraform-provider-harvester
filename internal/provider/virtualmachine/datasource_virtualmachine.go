package virtualmachine

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func DataSourceVirtualMachine() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVirtualMachineRead,
		Schema:      DataSourceSchema(),
	}
}

func dataSourceVirtualMachineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace := d.Get(constants.FieldCommonNamespace).(string)
	name := d.Get(constants.FieldCommonName).(string)
	vm, err := c.HarvesterClient.KubevirtV1().VirtualMachines(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	vmi, err := c.HarvesterClient.KubevirtV1().VirtualMachineInstances(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return diag.FromErr(err)
		}
		vmi = nil
	}
	return diag.FromErr(resourceVirtualMachineImport(d, vm, vmi, ""))
}
