package schedulebackup

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func DataSourceScheduleBackup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScheduleBackupRead,
		Schema:      DataSourceSchema(),
	}
}

func dataSourceScheduleBackupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}
	namespace := d.Get(constants.FieldCommonNamespace).(string)
	name := d.Get(constants.FieldCommonName).(string)

	scheduleVMBackup, err := c.HarvesterClient.HarvesterhciV1beta1().ScheduleVMBackups(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	vmName := scheduleVMBackup.Spec.VMBackupSpec.Source.Name
	d.SetId(fmt.Sprintf("%s/%s/%s", namespace, vmName, name))

	if err := d.Set(constants.FieldCommonNamespace, namespace); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(constants.FieldCommonName, name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(constants.FieldScheduleBackupVMName, helper.BuildNamespacedName(namespace, vmName)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(constants.FieldScheduleBackupSchedule, scheduleVMBackup.Spec.Cron); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(constants.FieldScheduleBackupRetain, scheduleVMBackup.Spec.Retain); err != nil {
		return diag.FromErr(err)
	}
	if len(scheduleVMBackup.Labels) > 0 {
		if err := d.Set(constants.FieldScheduleBackupLabels, scheduleVMBackup.Labels); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set(constants.FieldScheduleBackupEnabled, !scheduleVMBackup.Spec.Suspend); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
