package kubeovn_iptables_fip_rule

import (
	"testing"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// kube-ovn refuses any spec change once the object is ready ("not support
// change"), so every configurable spec field must force a replacement.
// Metadata (name, labels, tags, description) stays updatable in place.
func TestSpecFieldsForceNew(t *testing.T) {
	metadata := map[string]bool{
		constants.FieldCommonName:        true,
		constants.FieldCommonNamespace:   true,
		constants.FieldCommonDescription: true,
		constants.FieldCommonTags:        true,
		constants.FieldCommonLabels:      true,
	}
	for key, field := range Schema() {
		if metadata[key] || (field.Computed && !field.Optional && !field.Required) {
			continue
		}
		if !field.ForceNew {
			t.Errorf("spec field %q must be ForceNew", key)
		}
	}
}
