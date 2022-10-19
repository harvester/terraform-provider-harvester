package util

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/harvester/terraform-provider-harvester/pkg/importer"
)

func ResourceStatesSet(d *schema.ResourceData, getter *importer.StateGetter) error {
	for key, value := range getter.States {
		if err := d.Set(key, value); err != nil {
			return err
		}
	}
	d.SetId(getter.ID)
	return nil
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

func WatchOptions(name string, timeoutSeconds int64) metav1.ListOptions {
	return metav1.ListOptions{
		FieldSelector:  fmt.Sprintf("metadata.name=%s", name),
		Watch:          true,
		TimeoutSeconds: &timeoutSeconds,
	}
}
