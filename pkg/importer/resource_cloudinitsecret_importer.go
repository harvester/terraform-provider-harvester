package importer

import (
	"encoding/base64"

	corev1 "k8s.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceCloudInitSecretStateGetter(obj *corev1.Secret) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonNamespace:                  obj.Namespace,
		constants.FieldCommonName:                       obj.Name,
		constants.FieldCommonDescription:                GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                       GetTags(obj.Labels),
		constants.FieldCommonLabels:                     GetLabels(obj.Labels),
		constants.FieldCloudInitSecretUserDataBase64:    base64.StdEncoding.EncodeToString(obj.Data[constants.SecretDataKeyUserData]),
		constants.FieldCloudInitSecretNetworkDataBase64: base64.StdEncoding.EncodeToString(obj.Data[constants.SecretDataKeyNetworkData]),
	}

	return &StateGetter{
		ID:           helper.BuildID(obj.Namespace, obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeCloudInitSecret,
		States:       states,
	}, nil
}
