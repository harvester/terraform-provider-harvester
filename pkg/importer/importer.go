package importer

import (
	"regexp"
	"strings"

	"github.com/harvester/harvester/pkg/builder"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
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

// vmDeleteProtectionValue matches the label values that enable delete
// protection: only "true" and "True", as in KubeVirt and the Harvester webhook.
var vmDeleteProtectionValue = regexp.MustCompile(`^(true|True)$`)

// IsVMDeleteProtected reports whether the KubeVirt delete protection label
// enables protection, using the same check as the Harvester webhook.
func IsVMDeleteProtected(labels map[string]string) bool {
	return vmDeleteProtectionValue.MatchString(labels[constants.LabelVMDeleteProtection])
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
