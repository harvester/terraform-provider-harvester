package volumebackup

import (
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

var (
	_ util.Constructor = &Constructor{}
)

type Constructor struct {
	Volume *corev1.PersistentVolumeClaim
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.Volume.Labels).
		Labels(&c.Volume.Labels).
		Description(&c.Volume.Annotations)

	customProcessors := []util.Processor{
		{
			Field: constants.FieldVolumeBackupVolumeName,
			Parser: func(i interface{}) error {
				volumeName := i.(string)
				namespace, name, err := helper.NamespacedNamePartsByDefault(volumeName, c.Volume.Namespace)
				if err != nil {
					return err
				}
				// Store the volume reference in annotations for later use
				c.Volume.Annotations["terraform-provider-harvester/backup-volume"] = helper.BuildNamespacedName(namespace, name)
				return nil
			},
			Required: true,
		},
		{
			Field: constants.FieldVolumeBackupSchedule,
			Parser: func(i interface{}) error {
				schedule := i.(string)
				// Create the recurring job spec
				jobSpec := RecurringJobSpec{
					Name:        c.Volume.Name + "-backup",
					Task:        "backup",
					Cron:        schedule,
					Retain:      5,
					Concurrency: 1,
				}

				// Store in annotation
				if c.Volume.Annotations == nil {
					c.Volume.Annotations = make(map[string]string)
				}
				jobJSON, err := json.Marshal(jobSpec)
				if err != nil {
					return fmt.Errorf("failed to marshal backup job spec: %w", err)
				}
				c.Volume.Annotations[constants.AnnotationRecurringJobBackup] = string(jobJSON)
				return nil
			},
			Required: true,
		},
		{
			Field: constants.FieldVolumeBackupRetain,
			Parser: func(i interface{}) error {
				retain := i.(int)
				// Update the existing backup job spec
				if backupJSON := c.Volume.Annotations[constants.AnnotationRecurringJobBackup]; backupJSON != "" {
					var jobSpec RecurringJobSpec
					if err := json.Unmarshal([]byte(backupJSON), &jobSpec); err != nil {
						return fmt.Errorf("failed to unmarshal backup job spec: %w", err)
					}
					jobSpec.Retain = retain
					jobJSON, err := json.Marshal(jobSpec)
					if err != nil {
						return fmt.Errorf("failed to marshal backup job spec: %w", err)
					}
					c.Volume.Annotations[constants.AnnotationRecurringJobBackup] = string(jobJSON)
				}
				return nil
			},
		},
		{
			Field: constants.FieldVolumeBackupConcurrency,
			Parser: func(i interface{}) error {
				concurrency := i.(int)
				// Update the existing backup job spec
				if backupJSON := c.Volume.Annotations[constants.AnnotationRecurringJobBackup]; backupJSON != "" {
					var jobSpec RecurringJobSpec
					if err := json.Unmarshal([]byte(backupJSON), &jobSpec); err != nil {
						return fmt.Errorf("failed to unmarshal backup job spec: %w", err)
					}
					jobSpec.Concurrency = concurrency
					jobJSON, err := json.Marshal(jobSpec)
					if err != nil {
						return fmt.Errorf("failed to marshal backup job spec: %w", err)
					}
					c.Volume.Annotations[constants.AnnotationRecurringJobBackup] = string(jobJSON)
				}
				return nil
			},
		},
		{
			Field: constants.FieldVolumeBackupLabels,
			Parser: func(i interface{}) error {
				labels := i.(map[string]interface{})
				labelMap := make(map[string]string)
				for k, v := range labels {
					labelMap[k] = v.(string)
				}
				// Update the existing backup job spec
				if backupJSON := c.Volume.Annotations[constants.AnnotationRecurringJobBackup]; backupJSON != "" {
					var jobSpec RecurringJobSpec
					if err := json.Unmarshal([]byte(backupJSON), &jobSpec); err != nil {
						return fmt.Errorf("failed to unmarshal backup job spec: %w", err)
					}
					jobSpec.Labels = labelMap
					jobJSON, err := json.Marshal(jobSpec)
					if err != nil {
						return fmt.Errorf("failed to marshal backup job spec: %w", err)
					}
					c.Volume.Annotations[constants.AnnotationRecurringJobBackup] = string(jobJSON)
				}
				return nil
			},
		},
		{
			Field: constants.FieldVolumeBackupGroups,
			Parser: func(i interface{}) error {
				groups := i.([]interface{})
				groupList := make([]string, len(groups))
				for i, g := range groups {
					groupList[i] = g.(string)
				}
				// Update the existing backup job spec
				if backupJSON := c.Volume.Annotations[constants.AnnotationRecurringJobBackup]; backupJSON != "" {
					var jobSpec RecurringJobSpec
					if err := json.Unmarshal([]byte(backupJSON), &jobSpec); err != nil {
						return fmt.Errorf("failed to unmarshal backup job spec: %w", err)
					}
					jobSpec.Groups = groupList
					jobJSON, err := json.Marshal(jobSpec)
					if err != nil {
						return fmt.Errorf("failed to marshal backup job spec: %w", err)
					}
					c.Volume.Annotations[constants.AnnotationRecurringJobBackup] = string(jobJSON)
				}
				return nil
			},
		},
		{
			Field: constants.FieldVolumeBackupEnabled,
			Parser: func(i interface{}) error {
				enabled := i.(bool)
				if !enabled {
					// Remove the backup annotation if disabled
					delete(c.Volume.Annotations, constants.AnnotationRecurringJobBackup)
				}
				// If enabled, the schedule field will have already set the annotation
				return nil
			},
		},
	}
	return append(processors, customProcessors...)
}

func (c *Constructor) Validate() error {
	// Validate that the volume exists and can be accessed
	// This will be done in the resource create/update functions
	return nil
}

func (c *Constructor) Result() (interface{}, error) {
	return c.Volume, nil
}

func newVolumeBackupConstructor(volume *corev1.PersistentVolumeClaim) util.Constructor {
	return &Constructor{
		Volume: volume,
	}
}

func Creator(namespace, name string, volumeName string) util.Constructor {
	// Parse volume name to get namespace and name
	volNamespace, volName, err := helper.NamespacedNamePartsByDefault(volumeName, namespace)
	if err != nil {
		// This will be caught during validation
		volNamespace = namespace
		volName = volumeName
	}

	// Create a placeholder PVC that will be used to store backup configuration
	// The actual volume will be updated with annotations
	volume := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   volNamespace,
			Name:        volName,
			Annotations: make(map[string]string),
		},
	}
	return newVolumeBackupConstructor(volume)
}

func Updater(volume *corev1.PersistentVolumeClaim) util.Constructor {
	return newVolumeBackupConstructor(volume)
}
