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
		constants.FieldCommonNamespace:   true,
		constants.FieldCommonDescription: true,
		constants.FieldCommonTags:        true,
		constants.FieldCommonLabels:      true,
	}
	resources := Provider().ResourcesMap

	testcases := []string{
		constants.ResourceTypeKubeOVNIptablesEIP,
		constants.ResourceTypeKubeOVNIptablesDnatRule,
		constants.ResourceTypeKubeOVNIptablesSnatRule,
		constants.ResourceTypeKubeOVNIptablesFIPRule,
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

// Fields the kube-ovn controller stores but never applies to the running
// object must force a replacement; the others are reconciled in place.
func TestKubeOVNReplacedFields(t *testing.T) {
	resources := Provider().ResourcesMap

	testcases := []struct {
		resourceType string
		field        string
		forceNew     bool
	}{
		{constants.ResourceTypeKubeOVNVpcNatGateway, constants.FieldKubeOVNVpcNatGwVpc, true},
		{constants.ResourceTypeKubeOVNVpcNatGateway, constants.FieldKubeOVNVpcNatGwSubnet, true},
		{constants.ResourceTypeKubeOVNVpcNatGateway, constants.FieldKubeOVNVpcNatGwLanIP, true},
		{constants.ResourceTypeKubeOVNVpcNatGateway, constants.FieldKubeOVNVpcNatGwExternalSubnets, false},
		{constants.ResourceTypeKubeOVNVpcNatGateway, constants.FieldKubeOVNVpcNatGwSelector, false},
		{constants.ResourceTypeKubeOVNVpcNatGateway, constants.FieldKubeOVNVpcNatGwQoSPolicy, false},
		{constants.ResourceTypeKubeOVNQoSPolicy, constants.FieldKubeOVNQoSShared, true},
		{constants.ResourceTypeKubeOVNQoSPolicy, constants.FieldKubeOVNQoSBindingType, true},
		{constants.ResourceTypeKubeOVNQoSPolicy, constants.FieldKubeOVNQoSBandwidthLimitRules, false},
	}

	for _, tc := range testcases {
		t.Run(tc.resourceType+"."+tc.field, func(t *testing.T) {
			field, ok := resources[tc.resourceType].Schema[tc.field]
			if !ok {
				t.Fatalf("field %q not found", tc.field)
			}
			if field.ForceNew != tc.forceNew {
				t.Errorf("ForceNew = %v, want %v", field.ForceNew, tc.forceNew)
			}
		})
	}
}
