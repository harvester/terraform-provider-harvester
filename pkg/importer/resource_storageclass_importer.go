package importer

import (
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
	}
	return &StateGetter{
		ID:           helper.BuildID("", obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeStorageClass,
		States:       states,
	}, nil
}
