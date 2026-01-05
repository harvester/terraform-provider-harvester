// Package volumebackup provides the Terraform resource for managing Harvester VM backup schedules.
// This resource creates and manages ScheduleVMBackup CRDs which enable recurring backups
// of entire VirtualMachines (all disks) in Harvester.
package volumebackup

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

// ResourceVolumeBackup returns the Terraform resource schema for harvester_volume_backup.
// This resource manages recurring VM backups using Harvester's ScheduleVMBackup CRD.
// Note: Despite the name "volume_backup", this resource manages VM-level backups (all disks).
func ResourceVolumeBackup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVolumeBackupCreate,
		ReadContext:   resourceVolumeBackupRead,
		DeleteContext: resourceVolumeBackupDelete,
		UpdateContext: resourceVolumeBackupUpdate,
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

// resourceVolumeBackupCreate creates a new ScheduleVMBackup resource in Harvester.
// It creates a VM-level backup schedule that backs up all disks of the specified VM.
// The resource ID format is: namespace/vmname/jobname
func resourceVolumeBackupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	namespace := d.Get(constants.FieldCommonNamespace).(string)
	name := d.Get(constants.FieldCommonName).(string)
	
	// Get VM name - prefer vm_name over deprecated volume_name for backward compatibility
	var vmName string
	var vmNamespace string
	
	if vmNameRaw, ok := d.GetOk(constants.FieldVolumeBackupVMName); ok {
		vmNamespace, vmName, err = helper.NamespacedNamePartsByDefault(vmNameRaw.(string), namespace)
		if err != nil {
			return diag.FromErr(fmt.Errorf("invalid VM name format: %w", err))
		}
	} else if volumeNameRaw, ok := d.GetOk(constants.FieldVolumeBackupVolumeName); ok {
		// Backward compatibility: find VM from volume
		volNamespace, volName, err := helper.NamespacedNamePartsByDefault(volumeNameRaw.(string), namespace)
		if err != nil {
			return diag.FromErr(fmt.Errorf("invalid volume name format: %w", err))
		}

		// Get the existing volume to find the VM
		_, err = c.KubeClient.CoreV1().PersistentVolumeClaims(volNamespace).Get(ctx, volName, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return diag.Errorf("volume %s/%s not found", volNamespace, volName)
			}
			return diag.FromErr(err)
		}

		// Find the VM that uses this volume
		vms, err := c.HarvesterClient.KubevirtV1().VirtualMachines(volNamespace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, vm := range vms.Items {
				for _, vol := range vm.Spec.Template.Spec.Volumes {
					if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName == volName {
						vmName = vm.Name
						vmNamespace = volNamespace
						break
					}
				}
				if vmName != "" {
					break
				}
			}
		}
		
		if vmName == "" {
			return diag.Errorf("no VirtualMachine found using volume %s/%s", volNamespace, volName)
		}
	} else {
		return diag.Errorf("either vm_name or volume_name must be specified")
	}

	// Verify VM exists
	_, err = c.HarvesterClient.KubevirtV1().VirtualMachines(vmNamespace).Get(ctx, vmName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return diag.Errorf("virtual machine %s/%s not found", vmNamespace, vmName)
		}
		return diag.FromErr(err)
	}

	// Get backup configuration from Terraform schema
	schedule := d.Get(constants.FieldVolumeBackupSchedule).(string)
	retain := d.Get(constants.FieldVolumeBackupRetain).(int)
	enabled := d.Get(constants.FieldVolumeBackupEnabled).(bool)
	maxFailure := 3 // Default value for maximum consecutive failures before suspending

	// If backup is disabled, just set the ID and return (no ScheduleVMBackup will be created)
	if !enabled {
		d.SetId(fmt.Sprintf("%s/%s/%s", vmNamespace, vmName, name))
		return resourceVolumeBackupRead(ctx, d, meta)
	}

	// Create Harvester ScheduleVMBackup CRD for VM-level backup
	// This creates a recurring backup schedule for the entire VM (all disks)
	// Harvester only allows one ScheduleVMBackup per VM, so we handle existing schedules
	apiGroup := "kubevirt.io"
	scheduleVMBackup := &harvsterv1.ScheduleVMBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: vmNamespace,
		},
		Spec: harvsterv1.ScheduleVMBackupSpec{
			Cron:      schedule,
			Retain:    retain,
			MaxFailure: maxFailure,
			Suspend:   false,
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
	if labels := d.Get(constants.FieldVolumeBackupLabels); labels != nil {
		labelMap := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			labelMap[k] = v.(string)
		}
		scheduleVMBackup.ObjectMeta.Labels = labelMap
	}

	// Create or update ScheduleVMBackup
	// Harvester only allows one ScheduleVMBackup per VM, so we need to handle existing schedules
	_, err = c.HarvesterClient.HarvesterhciV1beta1().ScheduleVMBackups(vmNamespace).Create(ctx, scheduleVMBackup, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			// ScheduleVMBackup with this name already exists, update it
			existing, getErr := c.HarvesterClient.HarvesterhciV1beta1().ScheduleVMBackups(vmNamespace).Get(ctx, name, metav1.GetOptions{})
			if getErr != nil {
				return diag.FromErr(fmt.Errorf("failed to get existing ScheduleVMBackup: %w", getErr))
			}
			// Preserve resourceVersion and UID for update
			scheduleVMBackup.ObjectMeta.ResourceVersion = existing.ObjectMeta.ResourceVersion
			scheduleVMBackup.ObjectMeta.UID = existing.ObjectMeta.UID
			_, err = c.HarvesterClient.HarvesterhciV1beta1().ScheduleVMBackups(vmNamespace).Update(ctx, scheduleVMBackup, metav1.UpdateOptions{})
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to update ScheduleVMBackup: %w", err))
			}
		} else if strings.Contains(err.Error(), "already has backup schedule") {
			// Harvester validation error: VM already has a backup schedule with a different name
			// Find the existing schedule for this VM and update it
			existingSchedules, listErr := c.HarvesterClient.HarvesterhciV1beta1().ScheduleVMBackups(vmNamespace).List(ctx, metav1.ListOptions{})
			if listErr == nil {
				for _, existingSchedule := range existingSchedules.Items {
					if existingSchedule.Spec.VMBackupSpec.Source.Name == vmName {
						// Update the existing schedule with new configuration
						scheduleVMBackup.ObjectMeta.Name = existingSchedule.ObjectMeta.Name
						scheduleVMBackup.ObjectMeta.ResourceVersion = existingSchedule.ObjectMeta.ResourceVersion
						scheduleVMBackup.ObjectMeta.UID = existingSchedule.ObjectMeta.UID
						_, err = c.HarvesterClient.HarvesterhciV1beta1().ScheduleVMBackups(vmNamespace).Update(ctx, scheduleVMBackup, metav1.UpdateOptions{})
						if err != nil {
							return diag.FromErr(fmt.Errorf("failed to update existing ScheduleVMBackup: %w", err))
						}
						// Use the existing schedule name for the resource ID
						name = existingSchedule.ObjectMeta.Name
						break
					}
				}
			}
		} else {
			return diag.FromErr(fmt.Errorf("failed to create ScheduleVMBackup: %w", err))
		}
	}

	// Set the resource ID (format: namespace/vmname/jobname)
	d.SetId(fmt.Sprintf("%s/%s/%s", vmNamespace, vmName, name))

	return resourceVolumeBackupRead(ctx, d, meta)
}

// resourceVolumeBackupUpdate updates an existing ScheduleVMBackup resource.
// It handles changes to schedule, retain count, enabled status, and labels.
// If the VM name changes, the resource ID will be updated accordingly.
func resourceVolumeBackupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	// Parse ID: format is namespace/vmname/jobname
	id := d.Id()
	parts := strings.Split(id, "/")
	if len(parts) < 3 {
		return diag.Errorf("invalid resource ID format: %s (expected namespace/vmname/jobname)", id)
	}
	
	vmNamespace := parts[0]
	vmName := parts[1]
	jobName := parts[2]

	// Get VM name from resource data (may have changed)
	var targetVMName string
	var targetVMNamespace string
	
	if vmNameRaw, ok := d.GetOk(constants.FieldVolumeBackupVMName); ok {
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
	schedule := d.Get(constants.FieldVolumeBackupSchedule).(string)
	retain := d.Get(constants.FieldVolumeBackupRetain).(int)
	enabled := d.Get(constants.FieldVolumeBackupEnabled).(bool)
	maxFailure := 3

	// Create or update ScheduleVMBackup
	apiGroup := "kubevirt.io"
	scheduleVMBackup := &harvsterv1.ScheduleVMBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: targetVMNamespace,
		},
		Spec: harvsterv1.ScheduleVMBackupSpec{
			Cron:      schedule,
			Retain:    retain,
			MaxFailure: maxFailure,
			Suspend:   !enabled,
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
	if labels := d.Get(constants.FieldVolumeBackupLabels); labels != nil {
		labelMap := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			labelMap[k] = v.(string)
		}
		scheduleVMBackup.ObjectMeta.Labels = labelMap
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
		scheduleVMBackup.ObjectMeta.ResourceVersion = existing.ObjectMeta.ResourceVersion
		scheduleVMBackup.ObjectMeta.UID = existing.ObjectMeta.UID
		_, err = c.HarvesterClient.HarvesterhciV1beta1().ScheduleVMBackups(targetVMNamespace).Update(ctx, scheduleVMBackup, metav1.UpdateOptions{})
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to update ScheduleVMBackup: %w", err))
		}
	}

	// Update ID if VM changed
	if targetVMNamespace != vmNamespace || targetVMName != vmName {
		d.SetId(fmt.Sprintf("%s/%s/%s", targetVMNamespace, targetVMName, jobName))
	}

	return resourceVolumeBackupRead(ctx, d, meta)
}

func resourceVolumeBackupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	// Parse resource ID: format is namespace/vmname/jobname
	id := d.Id()
	parts := strings.Split(id, "/")
	if len(parts) < 3 {
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
	if err := d.Set(constants.FieldVolumeBackupVMName, helper.BuildNamespacedName(vmNamespace, vmName)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(constants.FieldVolumeBackupSchedule, scheduleVMBackup.Spec.Cron); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(constants.FieldVolumeBackupRetain, scheduleVMBackup.Spec.Retain); err != nil {
		return diag.FromErr(err)
	}
	if scheduleVMBackup.ObjectMeta.Labels != nil && len(scheduleVMBackup.ObjectMeta.Labels) > 0 {
		if err := d.Set(constants.FieldVolumeBackupLabels, scheduleVMBackup.ObjectMeta.Labels); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set(constants.FieldVolumeBackupEnabled, !scheduleVMBackup.Spec.Suspend); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// resourceVolumeBackupDelete deletes a ScheduleVMBackup resource.
// It removes the ScheduleVMBackup CRD from Harvester, which will stop the recurring backups.
func resourceVolumeBackupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	// Parse resource ID: format is namespace/vmname/jobname
	id := d.Id()
	parts := strings.Split(id, "/")
	if len(parts) < 3 {
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

