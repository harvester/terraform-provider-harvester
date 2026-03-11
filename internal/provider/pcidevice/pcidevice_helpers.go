package pcidevice

import (
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
)

// GVRs for Harvester PCI device CRDs
var (
	PCIDeviceClaimGVR = k8sschema.GroupVersionResource{
		Group:    "devices.harvesterhci.io",
		Version:  "v1beta1",
		Resource: "pcideviceclaims",
	}
	PCIDeviceGVR = k8sschema.GroupVersionResource{
		Group:    "devices.harvesterhci.io",
		Version:  "v1beta1",
		Resource: "pcidevices",
	}
)

// getField extracts a string field from an unstructured map.
func getField(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}
