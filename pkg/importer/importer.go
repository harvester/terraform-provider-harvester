package importer

import (
	"strings"

	"github.com/harvester/harvester/pkg/builder"
)

type StateGetter struct {
	ID           string
	Name         string
	ResourceType string
	States       map[string]interface{}
}

func GetTags(labels map[string]string) map[string]string {
	tags := map[string]string{}
	for labelKey, labelValue := range labels {
		if strings.HasPrefix(labelKey, builder.LabelPrefixHarvesterTag) {
			tags[strings.TrimPrefix(labelKey, builder.LabelPrefixHarvesterTag)] = labelValue
		}
	}
	return tags
}

func GetLabels(labels map[string]string) map[string]string {
	nottags := map[string]string{}

	for key, value := range labels {
		// Labels with the prefix tags.harvesterhci.io/ are ignored here, because
		// they are already handled as "tags".
		// Labels with the prefix harvesterhci.io/ are ignored as well, because they
		// are automatically added and should never appear in the user-specified
		// `labels` blocks in the .tf files.
		if !strings.HasPrefix(key, builder.LabelPrefixHarvesterTag) &&
			!strings.HasPrefix(key, builder.LabelAnnotationPrefixHarvester) {
			nottags[key] = value
		}
	}
	return nottags
}

func GetDescriptions(annotations map[string]string) string {
	return annotations[builder.AnnotationKeyDescription]
}
