package importer

import (
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceStorageClassStateGetter(obj *storagev1.StorageClass) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonName:                       obj.Name,
		constants.FieldCommonDescription:                GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                       GetTags(obj.Labels),
		constants.FieldStorageClassVolumeProvisioner:    obj.Provisioner,
		constants.FieldStorageClassParameters:           obj.Parameters,
		constants.FieldStorageClassAllowVolumeExpansion: *obj.AllowVolumeExpansion,
		constants.FieldStorageClassReclaimPolicy:        string(*obj.ReclaimPolicy),
		constants.FieldStorageClassVolumeBindingMode:    string(*obj.VolumeBindingMode),
		constants.FieldStorageClassIsDefault:            obj.Annotations["storageclass.kubernetes.io/is-default-class"] == "true",
		constants.FieldStorageClassAllowedTopologies:    getAllowedTopologies(obj.AllowedTopologies),
	}
	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeStorageClass,
		States:       states,
	}, nil
}

func getAllowedTopologies(topologies []corev1.TopologySelectorTerm) []map[string]interface{} {
	if len(topologies) == 0 {
		return nil
	}

	result := make([]map[string]interface{}, len(topologies))
	for i, topology := range topologies {
		requirements := make([]map[string]interface{}, len(topology.MatchLabelExpressions))
		for j, requirement := range topology.MatchLabelExpressions {
			requirements[j] = map[string]interface{}{
				"key":    requirement.Key,
				"values": requirement.Values,
			}
		}
		result[i] = map[string]interface{}{
			"match_label_expressions": requirements,
		}
	}
	return result
}
