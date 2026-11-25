package util

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func ResourceStatesSet(d *schema.ResourceData, getter *importer.StateGetter) error {
	return statesSet(d, getter, true)
}
func DataSourceStatesSet(d *schema.ResourceData, getter *importer.StateGetter) error {
	return statesSet(d, getter, false)
}

func statesSet(d *schema.ResourceData, getter *importer.StateGetter, managedLabelsOnly bool) error {
	for key, value := range getter.States {
		if managedLabelsOnly && key == constants.FieldCommonLabels {
			value = managedLabels(d.Get(key), value)
		}
		if err := d.Set(key, value); err != nil {
			return err
		}
	}
	d.SetId(getter.ID)
	return nil
}
func managedLabels(known, actual interface{}) interface{} {
	knownLabels, ok := known.(map[string]interface{})
	if !ok {
		return actual
	}
	actualLabels, ok := actual.(map[string]string)
	if !ok {
		return actual
	}
	managed := map[string]string{}
	for key, value := range actualLabels {
		if _, found := knownLabels[key]; found {
			managed[key] = value
		}
	}
	return managed
}

func HasDeleted(events watch.Interface) bool {
	deleted := false
	for event := range events.ResultChan() {
		if event.Type == watch.Deleted {
			events.Stop()
			deleted = true
		}
	}
	return deleted
}

func WatchOptions(name string, timeout time.Duration) metav1.ListOptions {
	return metav1.ListOptions{
		FieldSelector:  fmt.Sprintf("metadata.name=%s", name),
		Watch:          true,
		TimeoutSeconds: func() *int64 { to := timeout.Milliseconds() / 1000; return &to }(),
	}
}
