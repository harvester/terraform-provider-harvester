package importer

import (
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

// getNamespaceLabels filters labels for namespaces, excluding
// auto-managed Kubernetes and Rancher labels in addition to the
// standard harvester tag/annotation prefixes.
func getNamespaceLabels(labels map[string]string) map[string]string {
	filtered := GetLabels(labels)
	for key := range filtered {
		if key == "kubernetes.io/metadata.name" ||
			strings.HasPrefix(key, "cattle.io/") ||
			strings.HasPrefix(key, "lifecycle.cattle.io/") {
			delete(filtered, key)
		}
	}
	return filtered
}

func ResourceNamespaceStateGetter(obj *corev1.Namespace) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:        obj.Name,
		constants.FieldCommonDescription: GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:        GetTags(obj.Labels),
		constants.FieldCommonLabels:      getNamespaceLabels(obj.Labels),
		constants.FieldCommonState:       string(obj.Status.Phase),
	}
	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeNamespace,
		States:       states,
	}, nil
}
