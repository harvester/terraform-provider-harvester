// Package importer provides import functionality for Harvester resources.
// This file handles importing existing ScheduleVMBackup resources into Terraform state.
package importer

import (
	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

// ResourceScheduleBackupStateGetter imports an existing ScheduleVMBackup CRD into Terraform state.
// It reads the ScheduleVMBackup resource and converts it to Terraform state format.
// The resource ID format is: namespace/vmname/jobname
func ResourceScheduleBackupStateGetter(obj *harvsterv1.ScheduleVMBackup) (*StateGetter, error) {
	// Extract VM name from the ScheduleVMBackup spec
	vmName := obj.Spec.VMBackupSpec.Source.Name
	if vmName == "" {
		// Invalid ScheduleVMBackup, skip it
		return nil, nil
	}

	// Build the resource states from the ScheduleVMBackup CRD
	states := map[string]interface{}{
		constants.FieldCommonNamespace:           obj.Namespace,
		constants.FieldCommonName:                obj.Name,
		constants.FieldScheduleBackupVMName:      helper.BuildNamespacedName(obj.Namespace, vmName),
		constants.FieldScheduleBackupSchedule:    obj.Spec.Cron,
		constants.FieldScheduleBackupRetain:      obj.Spec.Retain,
		constants.FieldScheduleBackupConcurrency: 1, // Default value, not used by ScheduleVMBackup
		constants.FieldScheduleBackupEnabled:     !obj.Spec.Suspend,
	}

	// Add labels if present
	if len(obj.Labels) > 0 {
		states[constants.FieldScheduleBackupLabels] = obj.Labels
	}

	// Build the resource ID: namespace/vmname/jobname
	resourceID := helper.BuildID(obj.Namespace, vmName)
	resourceID = resourceID + "/" + obj.Name

	return &StateGetter{
		ID:           resourceID,
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeScheduleBackup,
		States:       states,
	}, nil
}
