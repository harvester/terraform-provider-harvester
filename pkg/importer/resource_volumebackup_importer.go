// Package importer provides import functionality for Harvester resources.
// This file handles importing existing ScheduleVMBackup resources into Terraform state.
package importer

import (
	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

// ResourceVolumeBackupStateGetter imports an existing ScheduleVMBackup CRD into Terraform state.
// It reads the ScheduleVMBackup resource and converts it to Terraform state format.
// The resource ID format is: namespace/vmname/jobname
func ResourceVolumeBackupStateGetter(obj *harvsterv1.ScheduleVMBackup) (*StateGetter, error) {
	// Extract VM name from the ScheduleVMBackup spec
	vmName := obj.Spec.VMBackupSpec.Source.Name
	if vmName == "" {
		// Invalid ScheduleVMBackup, skip it
		return nil, nil
	}

	// Build the resource states from the ScheduleVMBackup CRD
	states := map[string]interface{}{
		constants.FieldCommonNamespace:        obj.Namespace,
		constants.FieldCommonName:             obj.Name,
		constants.FieldVolumeBackupVMName:      helper.BuildNamespacedName(obj.Namespace, vmName),
		constants.FieldVolumeBackupSchedule:   obj.Spec.Cron,
		constants.FieldVolumeBackupRetain:     obj.Spec.Retain,
		constants.FieldVolumeBackupConcurrency: 1, // Default value, not used by ScheduleVMBackup
		constants.FieldVolumeBackupEnabled:    !obj.Spec.Suspend,
	}

	// Add labels if present
	if obj.Labels != nil && len(obj.Labels) > 0 {
		states[constants.FieldVolumeBackupLabels] = obj.Labels
	}

	// Build the resource ID: namespace/vmname/jobname
	resourceID := helper.BuildID(obj.Namespace, vmName)
	resourceID = resourceID + "/" + obj.Name

	return &StateGetter{
		ID:           resourceID,
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeVolumeBackup,
		States:       states,
	}, nil
}

