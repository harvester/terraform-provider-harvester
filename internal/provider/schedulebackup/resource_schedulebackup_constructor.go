package schedulebackup

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

// getOrCreateJobSpec retrieves the existing job spec from annotations or creates a new one.
func (c *Constructor) getOrCreateJobSpec() (RecurringJobSpec, error) {
	if c.Volume.Annotations == nil {
		c.Volume.Annotations = make(map[string]string)
	}

	backupJSON := c.Volume.Annotations[constants.AnnotationRecurringJobBackup]
	if backupJSON != "" {
		var jobSpec RecurringJobSpec
		if err := json.Unmarshal([]byte(backupJSON), &jobSpec); err != nil {
			return RecurringJobSpec{}, fmt.Errorf("failed to unmarshal backup job spec: %w", err)
		}
		return jobSpec, nil
	}

	// Create new job spec
	return RecurringJobSpec{
		Name:        c.Volume.Name + "-backup",
		Task:        "backup",
		Retain:      5,
		Concurrency: 1,
	}, nil
}

// saveJobSpec saves the job spec to annotations.
func (c *Constructor) saveJobSpec(jobSpec RecurringJobSpec) error {
	if c.Volume.Annotations == nil {
		c.Volume.Annotations = make(map[string]string)
	}
	jobJSON, err := json.Marshal(jobSpec)
	if err != nil {
		return fmt.Errorf("failed to marshal backup job spec: %w", err)
	}
	c.Volume.Annotations[constants.AnnotationRecurringJobBackup] = string(jobJSON)
	return nil
}

// updateJobSpecField updates a specific field in the job spec.
func (c *Constructor) updateJobSpecField(updateFunc func(*RecurringJobSpec)) error {
	jobSpec, err := c.getOrCreateJobSpec()
	if err != nil {
		return err
	}
	updateFunc(&jobSpec)
	return c.saveJobSpec(jobSpec)
}

func (c *Constructor) Setup() util.Processors {
	processors := util.NewProcessors().
		Tags(&c.Volume.Labels).
		Labels(&c.Volume.Labels).
		Description(&c.Volume.Annotations)

	customProcessors := []util.Processor{
		{
			Field: constants.FieldScheduleBackupVolumeName,
			Parser: func(i interface{}) error {
				volumeName := i.(string)
				namespace, name, err := helper.NamespacedNamePartsByDefault(volumeName, c.Volume.Namespace)
				if err != nil {
					return err
				}
				if c.Volume.Annotations == nil {
					c.Volume.Annotations = make(map[string]string)
				}
				c.Volume.Annotations["terraform-provider-harvester/backup-volume"] = helper.BuildNamespacedName(namespace, name)
				return nil
			},
			Required: true,
		},
		{
			Field: constants.FieldScheduleBackupSchedule,
			Parser: func(i interface{}) error {
				schedule := i.(string)
				jobSpec, err := c.getOrCreateJobSpec()
				if err != nil {
					return err
				}
				jobSpec.Cron = schedule
				return c.saveJobSpec(jobSpec)
			},
			Required: true,
		},
		{
			Field: constants.FieldScheduleBackupRetain,
			Parser: func(i interface{}) error {
				retain := i.(int)
				return c.updateJobSpecField(func(js *RecurringJobSpec) {
					js.Retain = retain
				})
			},
		},
		{
			Field: constants.FieldScheduleBackupConcurrency,
			Parser: func(i interface{}) error {
				concurrency := i.(int)
				return c.updateJobSpecField(func(js *RecurringJobSpec) {
					js.Concurrency = concurrency
				})
			},
		},
		{
			Field: constants.FieldScheduleBackupLabels,
			Parser: func(i interface{}) error {
				labels := i.(map[string]interface{})
				labelMap := make(map[string]string)
				for k, v := range labels {
					labelMap[k] = v.(string)
				}
				return c.updateJobSpecField(func(js *RecurringJobSpec) {
					js.Labels = labelMap
				})
			},
		},
		{
			Field: constants.FieldScheduleBackupGroups,
			Parser: func(i interface{}) error {
				groups := i.([]interface{})
				groupList := make([]string, len(groups))
				for i, g := range groups {
					groupList[i] = g.(string)
				}
				return c.updateJobSpecField(func(js *RecurringJobSpec) {
					js.Groups = groupList
				})
			},
		},
		{
			Field: constants.FieldScheduleBackupEnabled,
			Parser: func(i interface{}) error {
				enabled := i.(bool)
				if !enabled {
					if c.Volume.Annotations != nil {
						delete(c.Volume.Annotations, constants.AnnotationRecurringJobBackup)
					}
				}
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

func newScheduleBackupConstructor(volume *corev1.PersistentVolumeClaim) util.Constructor {
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
	return newScheduleBackupConstructor(volume)
}

func Updater(volume *corev1.PersistentVolumeClaim) util.Constructor {
	return newScheduleBackupConstructor(volume)
}
