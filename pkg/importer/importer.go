package importer

import (
	"strings"

	"github.com/harvester/harvester/pkg/builder"

	"github.com/harvester/terraform-provider-harvester/pkg/helper"
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
		// Labels in the harvesterhci.io domain are ignored here. Those with the
		// prefix tag.harvesterhci.io/ are already handled as "tags", and the
		// remaining ones are maintained by Harvester and its controllers, so they
		// should never appear in the user-specified `labels` blocks in the .tf
		// files.
		if !helper.IsReservedLabelKey(key) {
			nottags[key] = value
		}
	}
	return nottags
}

func GetDescriptions(annotations map[string]string) string {
	return annotations[builder.AnnotationKeyDescription]
}
