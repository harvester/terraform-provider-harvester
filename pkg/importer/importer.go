package importer

import (
	"strings"

	"github.com/harvester/terraform-provider-harvester/pkg/builder"
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

func GetDescriptions(annotations map[string]string) string {
	return annotations[builder.AnnotationKeyDescription]
}
