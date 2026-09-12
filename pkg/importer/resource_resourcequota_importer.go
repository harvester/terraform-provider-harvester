package importer

import (
	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceResourceQuotaStateGetter(obj *harvsterv1.ResourceQuota) (*StateGetter, error) {
	vmQuotas := map[string]interface{}{}
	for k, v := range obj.Spec.SnapshotLimit.VMTotalSnapshotSizeQuota {
		vmQuotas[k] = int(v)
	}

	vmUsage := map[string]interface{}{}
	for k, v := range obj.Status.SnapshotLimitStatus.VMTotalSnapshotSizeUsage {
		vmUsage[k] = int(v)
	}

	states := map[string]interface{}{
		constants.FieldCommonNamespace:                              obj.Namespace,
		constants.FieldCommonName:                                   obj.Name,
		constants.FieldCommonDescription:                            GetDescriptions(obj.Annotations),
		constants.FieldCommonTags:                                   GetTags(obj.Labels),
		constants.FieldCommonLabels:                                 GetLabels(obj.Labels),
		constants.FieldResourceQuotaNamespaceTotalSnapshotSizeQuota: int(obj.Spec.SnapshotLimit.NamespaceTotalSnapshotSizeQuota),
		constants.FieldResourceQuotaVMTotalSnapshotSizeQuota:        vmQuotas,
		constants.FieldResourceQuotaNamespaceTotalSnapshotSizeUsage: int(obj.Status.SnapshotLimitStatus.NamespaceTotalSnapshotSizeUsage),
		constants.FieldResourceQuotaVMTotalSnapshotSizeUsage:        vmUsage,
	}

	return &StateGetter{
		ID:           helper.BuildID(obj.Namespace, obj.Name),
		Name:         obj.Name,
		ResourceType: constants.ResourceTypeResourceQuota,
		States:       states,
	}, nil
}
