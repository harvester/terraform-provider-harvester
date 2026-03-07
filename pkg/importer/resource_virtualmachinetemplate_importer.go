package importer

import (
	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceVirtualMachineTemplateStateGetter(obj *harvsterv1.VirtualMachineTemplate) (*StateGetter, error) {
	states := map[string]interface{}{
		constants.FieldCommonNamespace:                        obj.Namespace,
		constants.FieldCommonName:                             obj.Name,
		constants.FieldCommonDescription:                      GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                             GetTags(obj.Labels),
		constants.FieldCommonLabels:                           GetLabels(obj.Labels),
		constants.FieldVirtualMachineTemplateDefaultVersionID: obj.Spec.DefaultVersionID,
		constants.FieldVirtualMachineTemplateDefaultVersion:   obj.Status.DefaultVersion,
		constants.FieldVirtualMachineTemplateLatestVersion:    obj.Status.LatestVersion,
	}

	return &StateGetter{
		ID:           helper.BuildID(obj.Namespace, obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeVirtualMachineTemplate,
		States:       states,
	}, nil
}
