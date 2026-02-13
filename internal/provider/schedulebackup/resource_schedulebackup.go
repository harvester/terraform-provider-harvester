// Package schedulebackup provides the Terraform resource for managing Harvester VM backup schedules.
// This resource creates and manages ScheduleVMBackup CRDs which enable recurring backups
// of entire VirtualMachines (all disks) in Harvester.
package schedulebackup

import (
	"context"
	"fmt"
	"strings"
	"time"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

// ResourceScheduleBackup returns the Terraform resource schema for harvester_schedule_backup.
// This resource manages recurring VM backups using Harvester's ScheduleVMBackup CRD.
// This resource manages VM-level backup schedules (all disks).
func ResourceScheduleBackup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScheduleBackupCreate,
		ReadContext:   resourceScheduleBackupRead,
		DeleteContext: resourceScheduleBackupDelete,
		UpdateContext: resourceScheduleBackupUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: Schema(),
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(2 * time.Minute),
			Read:    schema.DefaultTimeout(2 * time.Minute),
			Update:  schema.DefaultTimeout(2 * time.Minute),
			Delete:  schema.DefaultTimeout(2 * time.Minute),
			Default: schema.DefaultTimeout(2 * time.Minute),
		},
	}
}

// getVMNameFromResource extracts the VM name and namespace from the Terraform resource data.
// It supports both vm_name (preferred) and deprecated volume_name for backward compatibility.
func getVMNameFromResource(ctx context.Context, c *client.Client, d *schema.ResourceData, namespace string) (vmNamespace, vmName string, diags diag.Diagnostics) {
	if vmNameRaw, ok := d.GetOk(constants.FieldScheduleBackupVMName); ok {
		var err error
		vmNamespace, vmName, err = helper.NamespacedNamePartsByDefault(vmNameRaw.(string), namespace)
		if err != nil {
			return "", "", diag.FromErr(fmt.Errorf("invalid VM name format: %w", err))
		}
		return vmNamespace, vmName, nil
	}

	if volumeNameRaw, ok := d.GetOk(constants.FieldScheduleBackupVolumeName); ok {
		return findVMFromVolume(ctx, c, volumeNameRaw.(string), namespace)
	}

	return "", "", diag.Errorf("either vm_name or volume_name must be specified")
}

// findVMFromVolume finds the VM that uses the specified volume (backward compatibility).
func findVMFromVolume(ctx context.Context, c *client.Client, volumeNameRaw, namespace string) (vmNamespace, vmName string, diags diag.Diagnostics) {
	volNamespace, volName, err := helper.NamespacedNamePartsByDefault(volumeNameRaw, namespace)
	if err != nil {
		return "", "", diag.FromErr(fmt.Errorf("invalid volume name format: %w", err))
	}

	// Verify volume exists
	_, err = c.KubeClient.CoreV1().PersistentVolumeClaims(volNamespace).Get(ctx, volName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return "", "", diag.Errorf("volume %s/%s not found", volNamespace, volName)
		}
		return "", "", diag.FromErr(err)
	}

	// Find the VM that uses this volume
	vms, err := c.HarvesterClient.KubevirtV1().VirtualMachines(volNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", "", diag.FromErr(fmt.Errorf("failed to list VMs: %w", err))
	}

	for _, vm := range vms.Items {
		for _, vol := range vm.Spec.Template.Spec.Volumes {
			if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName == volName {
				return volNamespace, vm.Name, nil
			}
		}
	}

	return "", "", diag.Errorf("no VirtualMachine found using volume %s/%s", volNamespace, volName)
}

// buildScheduleVMBackup creates a ScheduleVMBackup object from Terraform resource data.
func buildScheduleVMBackup(vmNamespace, vmName, name, schedule string, retain int, labels map[string]interface{}) *harvsterv1.ScheduleVMBackup {
	apiGroup := "kubevirt.io"
	// MaxFailure must be less than Retain per Harvester webhook validation
	maxFailure := retain - 1
	if maxFailure < 1 {
		maxFailure = 1
	}
	scheduleVMBackup := &harvsterv1.ScheduleVMBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: vmNamespace,
		},
		Spec: harvsterv1.ScheduleVMBackupSpec{
			Cron:       schedule,
			Retain:     retain,
			MaxFailure: maxFailure,
			Suspend:    false,
			VMBackupSpec: harvsterv1.VirtualMachineBackupSpec{
				Type: "backup",
				Source: corev1.TypedLocalObjectReference{
					Kind:     "VirtualMachine",
					Name:     vmName,
					APIGroup: &apiGroup,
				},
			},
		},
	}

	// Add labels if specified
	if labels != nil {
		labelMap := make(map[string]string)
		for k, v := range labels {
			labelMap[k] = v.(string)
		}
		scheduleVMBackup.Labels = labelMap
	}

	return scheduleVMBackup
}

// createOrUpdateScheduleVMBackup creates or updates a ScheduleVMBackup resource.
// It handles the case where a schedule already exists for the VM.
func createOrUpdateScheduleVMBackup(ctx context.Context, c *client.Client, scheduleVMBackup *harvsterv1.ScheduleVMBackup, vmNamespace, vmName string) (jobName string, diags diag.Diagnostics) {
	name := scheduleVMBackup.Name
	_, err := c.HarvesterClient.HarvesterhciV1beta1().ScheduleVMBackups(vmNamespace).Create(ctx, scheduleVMBackup, metav1.CreateOptions{})
	if err == nil {
		return name, nil
	}

	if apierrors.IsAlreadyExists(err) {
		// ScheduleVMBackup with this name already exists, update it
		existing, getErr := c.HarvesterClient.HarvesterhciV1beta1().ScheduleVMBackups(vmNamespace).Get(ctx, name, metav1.GetOptions{})
		if getErr != nil {
			return "", diag.FromErr(fmt.Errorf("failed to get existing ScheduleVMBackup: %w", getErr))
		}
		scheduleVMBackup.ResourceVersion = existing.ResourceVersion
		scheduleVMBackup.UID = existing.UID
		_, err = c.HarvesterClient.HarvesterhciV1beta1().ScheduleVMBackups(vmNamespace).Update(ctx, scheduleVMBackup, metav1.UpdateOptions{})
		if err != nil {
			return "", diag.FromErr(fmt.Errorf("failed to update ScheduleVMBackup: %w", err))
		}
		return name, nil
	}

	if strings.Contains(err.Error(), "already has backup schedule") {
		// Find and update existing schedule for this VM
		return updateExistingScheduleForVM(ctx, c, scheduleVMBackup, vmNamespace, vmName)
	}

	return "", diag.FromErr(fmt.Errorf("failed to create ScheduleVMBackup: %w", err))
}

// updateExistingScheduleForVM finds and updates an existing ScheduleVMBackup for the specified VM.
func updateExistingScheduleForVM(ctx context.Context, c *client.Client, scheduleVMBackup *harvsterv1.ScheduleVMBackup, vmNamespace, vmName string) (jobName string, diags diag.Diagnostics) {
	existingSchedules, listErr := c.HarvesterClient.HarvesterhciV1beta1().ScheduleVMBackups(vmNamespace).List(ctx, metav1.ListOptions{})
	if listErr != nil {
		return "", diag.FromErr(fmt.Errorf("failed to list existing schedules: %w", listErr))
	}

	for _, existingSchedule := range existingSchedules.Items {
		if existingSchedule.Spec.VMBackupSpec.Source.Name == vmName {
			scheduleVMBackup.Name = existingSchedule.Name
			scheduleVMBackup.ResourceVersion = existingSchedule.ResourceVersion
			scheduleVMBackup.UID = existingSchedule.UID
			_, err := c.HarvesterClient.HarvesterhciV1beta1().ScheduleVMBackups(vmNamespace).Update(ctx, scheduleVMBackup, metav1.UpdateOptions{})
			if err != nil {
				return "", diag.FromErr(fmt.Errorf("failed to update existing ScheduleVMBackup: %w", err))
			}
			return existingSchedule.Name, nil
		}
	}

	return "", diag.Errorf("failed to find existing schedule for VM %s/%s", vmNamespace, vmName)
}

// resourceScheduleBackupCreate creates a new ScheduleVMBackup resource in Harvester.
// It creates a VM-level backup schedule that backs up all disks of the specified VM.
// The resource ID format is: namespace/vmname/jobname
func resourceScheduleBackupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	namespace := d.Get(constants.FieldCommonNamespace).(string)
	name := d.Get(constants.FieldCommonName).(string)

	// Get VM name and namespace
	vmNamespace, vmName, diags := getVMNameFromResource(ctx, c, d, namespace)
	if diags != nil {
		return diags
	}

	// Verify VM exists
	_, err = c.HarvesterClient.KubevirtV1().VirtualMachines(vmNamespace).Get(ctx, vmName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return diag.Errorf("virtual machine %s/%s not found", vmNamespace, vmName)
		}
		return diag.FromErr(err)
	}

	// Get backup configuration
	schedule := d.Get(constants.FieldScheduleBackupSchedule).(string)
	retain := d.Get(constants.FieldScheduleBackupRetain).(int)
	enabled := d.Get(constants.FieldScheduleBackupEnabled).(bool)
	var labelMap map[string]interface{}
	if labels, ok := d.GetOk(constants.FieldScheduleBackupLabels); ok {
		labelMap = labels.(map[string]interface{})
	}

	// Build ScheduleVMBackup object
	scheduleVMBackup := buildScheduleVMBackup(vmNamespace, vmName, name, schedule, retain, labelMap)
	if !enabled {
		scheduleVMBackup.Spec.Suspend = true
	}

	// Create or update ScheduleVMBackup
	jobName, diags := createOrUpdateScheduleVMBackup(ctx, c, scheduleVMBackup, vmNamespace, vmName)
	if diags != nil {
		return diags
	}

	// Set the resource ID (format: namespace/vmname/jobname)
	d.SetId(fmt.Sprintf("%s/%s/%s", vmNamespace, vmName, jobName))

	return resourceScheduleBackupRead(ctx, d, meta)
}

// resourceScheduleBackupUpdate updates an existing ScheduleVMBackup resource.
// It handles changes to schedule, retain count, enabled status, and labels.
// If the VM name changes, the resource ID will be updated accordingly.
func resourceScheduleBackupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	// Parse ID: format is namespace/vmname/jobname
	id := d.Id()
	parts := strings.Split(id, "/")
	if len(parts) != 3 {
		return diag.Errorf("invalid resource ID format: %s (expected namespace/vmname/jobname)", id)
	}

	vmNamespace := parts[0]
	vmName := parts[1]
	jobName := parts[2]

	// Get VM name from resource data (may have changed)
	var targetVMName string
	var targetVMNamespace string

	if vmNameRaw, ok := d.GetOk(constants.FieldScheduleBackupVMName); ok {
		targetVMNamespace, targetVMName, err = helper.NamespacedNamePartsByDefault(vmNameRaw.(string), vmNamespace)
		if err != nil {
			return diag.FromErr(fmt.Errorf("invalid VM name format: %w", err))
		}
	} else {
		// Use existing VM from ID
		targetVMNamespace = vmNamespace
		targetVMName = vmName
	}

	// Verify VM exists
	_, err = c.HarvesterClient.KubevirtV1().VirtualMachines(targetVMNamespace).Get(ctx, targetVMName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// Get backup configuration
	schedule := d.Get(constants.FieldScheduleBackupSchedule).(string)
	retain := d.Get(constants.FieldScheduleBackupRetain).(int)
	enabled := d.Get(constants.FieldScheduleBackupEnabled).(bool)
	// MaxFailure must be less than Retain per Harvester webhook validation
	maxFailure := retain - 1
	if maxFailure < 1 {
		maxFailure = 1
	}

	// Create or update ScheduleVMBackup
	apiGroup := "kubevirt.io"
	scheduleVMBackup := &harvsterv1.ScheduleVMBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: targetVMNamespace,
		},
		Spec: harvsterv1.ScheduleVMBackupSpec{
			Cron:       schedule,
			Retain:     retain,
			MaxFailure: maxFailure,
			Suspend:    !enabled,
			VMBackupSpec: harvsterv1.VirtualMachineBackupSpec{
				Type: "backup",
				Source: corev1.TypedLocalObjectReference{
					Kind:     "VirtualMachine",
					Name:     targetVMName,
					APIGroup: &apiGroup,
				},
			},
		},
	}

	// Add labels if specified
	if labels, ok := d.GetOk(constants.FieldScheduleBackupLabels); ok {
		labelMap := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			labelMap[k] = v.(string)
		}
		scheduleVMBackup.Labels = labelMap
	}

	existing, getErr := c.HarvesterClient.HarvesterhciV1beta1().ScheduleVMBackups(targetVMNamespace).Get(ctx, jobName, metav1.GetOptions{})
	if getErr != nil {
		if apierrors.IsNotFound(getErr) {
			// Create if not exists
			_, err = c.HarvesterClient.HarvesterhciV1beta1().ScheduleVMBackups(targetVMNamespace).Create(ctx, scheduleVMBackup, metav1.CreateOptions{})
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to create ScheduleVMBackup: %w", err))
			}
		} else {
			return diag.FromErr(fmt.Errorf("failed to get ScheduleVMBackup: %w", getErr))
		}
	} else {
		// Update existing
		scheduleVMBackup.ResourceVersion = existing.ResourceVersion
		scheduleVMBackup.UID = existing.UID
		_, err = c.HarvesterClient.HarvesterhciV1beta1().ScheduleVMBackups(targetVMNamespace).Update(ctx, scheduleVMBackup, metav1.UpdateOptions{})
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to update ScheduleVMBackup: %w", err))
		}
	}

	// Update ID if VM changed
	if targetVMNamespace != vmNamespace || targetVMName != vmName {
		d.SetId(fmt.Sprintf("%s/%s/%s", targetVMNamespace, targetVMName, jobName))
	}

	return resourceScheduleBackupRead(ctx, d, meta)
}

func resourceScheduleBackupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	// Parse resource ID: format is namespace/vmname/jobname
	id := d.Id()
	parts := strings.Split(id, "/")
	if len(parts) != 3 {
		return diag.Errorf("invalid resource ID format: %s (expected namespace/vmname/jobname)", id)
	}

	vmNamespace := parts[0]
	vmName := parts[1]
	jobName := parts[2]

	// Get the ScheduleVMBackup CRD
	scheduleVMBackup, err := c.HarvesterClient.HarvesterhciV1beta1().ScheduleVMBackups(vmNamespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Resource doesn't exist, mark as removed
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// Verify the ScheduleVMBackup is for the correct VM (safety check)
	if scheduleVMBackup.Spec.VMBackupSpec.Source.Name != vmName {
		// VM mismatch, mark resource as removed
		d.SetId("")
		return nil
	}

	// Set the resource data
	if err := d.Set(constants.FieldCommonNamespace, vmNamespace); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(constants.FieldCommonName, jobName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(constants.FieldScheduleBackupVMName, helper.BuildNamespacedName(vmNamespace, vmName)); err != nil {
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

// resourceScheduleBackupDelete deletes a ScheduleVMBackup resource.
// It removes the ScheduleVMBackup CRD from Harvester, which will stop the recurring backups.
func resourceScheduleBackupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	// Parse resource ID: format is namespace/vmname/jobname
	id := d.Id()
	parts := strings.Split(id, "/")
	if len(parts) != 3 {
		return diag.Errorf("invalid resource ID format: %s (expected namespace/vmname/jobname)", id)
	}

	vmNamespace := parts[0]
	jobName := parts[2]

	// Delete the Harvester ScheduleVMBackup CRD
	// This will stop the recurring backups for the VM
	err = c.HarvesterClient.HarvesterhciV1beta1().ScheduleVMBackups(vmNamespace).Delete(ctx, jobName, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return diag.FromErr(fmt.Errorf("failed to delete ScheduleVMBackup: %w", err))
	}

	// Mark resource as deleted
	d.SetId("")
	return nil
}
