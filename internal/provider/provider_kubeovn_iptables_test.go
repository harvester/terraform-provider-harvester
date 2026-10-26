package provider

import (
	"testing"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// kube-ovn refuses any spec change on an iptables NAT object once it is ready
// ("not support change"), so every configurable spec field must force a
// replacement. Metadata (name, labels, tags, description) stays updatable in
// place.
func TestKubeOVNIptablesSpecFieldsForceNew(t *testing.T) {
	metadata := map[string]bool{
		constants.FieldCommonName:        true,
		constants.FieldCommonDescription: true,
		constants.FieldCommonTags:        true,
		constants.FieldCommonLabels:      true,
	}
	resources := Provider().ResourcesMap

	testcases := []string{
		constants.ResourceTypeKubeOVNIptablesEIP,
		constants.ResourceTypeKubeOVNIptablesDnatRule,
		constants.ResourceTypeKubeOVNIptablesSnatRule,
	}

	for _, resourceType := range testcases {
		t.Run(resourceType, func(t *testing.T) {
			for key, field := range resources[resourceType].Schema {
				if metadata[key] || (field.Computed && !field.Optional && !field.Required) {
					continue
				}
				if !field.ForceNew {
					t.Errorf("spec field %q must be ForceNew", key)
				}
			}
		})
	}
}
