package importer

import (
	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	corev1 "k8s.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceKeyPairStateGetter(obj *harvsterv1.KeyPair) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonNamespace:    obj.Namespace,
		constants.FieldCommonName:         obj.Name,
		constants.FieldCommonDescription:  GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:         GetTags(obj.Labels),
		constants.FieldKeyPairPublicKey:   obj.Spec.PublicKey,
		constants.FieldKeyPairFingerPrint: obj.Status.FingerPrint,
	}

	states[constants.FieldCommonState] = constants.StateKeyPairNotValidated
	for _, condition := range obj.Status.Conditions {
		if condition.Type == harvsterv1.KeyPairValidated && condition.Status == corev1.ConditionTrue {
			states[constants.FieldCommonState] = constants.StateKeyPairValidated
		}
	}
	return &StateGetter{
		ID:           helper.BuildID(obj.Namespace, obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeKeyPair,
		States:       states,
	}, nil
}
