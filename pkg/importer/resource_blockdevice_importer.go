package importer

import (
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func ResourceBlockDeviceStateGetter(obj *unstructured.Unstructured) (*StateGetter, error) {
	namespace := obj.GetNamespace()
	name := obj.GetName()

	states := map[string]interface{}{
		constants.FieldCommonNamespace:   namespace,
		constants.FieldCommonName:        name,
		constants.FieldCommonDescription: GetDescriptions(obj.GetAnnotations()),
		constants.FieldCommonTags:        GetTags(obj.GetLabels()),
		constants.FieldCommonLabels:      getBlockDeviceLabels(obj.GetLabels()),
	}

	// Spec fields
	nodeName, _, _ := unstructured.NestedString(obj.Object, "spec", "nodeName")
	states[constants.FieldBlockDeviceNodeName] = nodeName

	devPath, _, _ := unstructured.NestedString(obj.Object, "spec", "devPath")
	states[constants.FieldBlockDeviceDevPath] = devPath

	provision, _, _ := unstructured.NestedBool(obj.Object, "spec", "provision")
	states[constants.FieldBlockDeviceProvision] = provision

	forceFormatted, _, _ := unstructured.NestedBool(obj.Object, "spec", "fileSystem", "forceFormatted")
	states[constants.FieldBlockDeviceForceFormatted] = forceFormatted

	// Device tags (spec.tags)
	specTags, _, _ := unstructured.NestedStringSlice(obj.Object, "spec", "tags")
	if specTags != nil {
		tagList := make([]interface{}, len(specTags))
		for i, t := range specTags {
			tagList[i] = t
		}
		states[constants.FieldBlockDeviceDeviceTags] = tagList
	} else {
		states[constants.FieldBlockDeviceDeviceTags] = []interface{}{}
	}

	// Provisioner
	states[constants.FieldBlockDeviceProvisioner] = readProvisioner(obj)

	// Status fields
	provisionPhase, _, _ := unstructured.NestedString(obj.Object, "status", "provisionPhase")
	states[constants.FieldBlockDeviceProvisionPhase] = provisionPhase

	state, _, _ := unstructured.NestedString(obj.Object, "status", "state")
	states[constants.FieldCommonState] = state

	// Device status
	states[constants.FieldBlockDeviceDeviceStatus] = readDeviceStatus(obj)

	return &StateGetter{
		ID:           helper.BuildID(namespace, name),
		Name:         name,
		ResourceType: constants.ResourceTypeBlockDevice,
		States:       states,
	}, nil
}

func readProvisioner(obj *unstructured.Unstructured) []interface{} {
	prov, found, _ := unstructured.NestedMap(obj.Object, "spec", "provisioner")
	if !found || len(prov) == 0 {
		return []interface{}{}
	}

	result := map[string]interface{}{
		constants.FieldBlockDeviceProvisionerLonghorn: []interface{}{},
		constants.FieldBlockDeviceProvisionerLVM:      []interface{}{},
	}

	if lh, ok := prov["longhorn"]; ok {
		if lhMap, ok := lh.(map[string]interface{}); ok {
			engineVersion, _ := lhMap["engineVersion"].(string)
			diskDriver, _ := lhMap["diskDriver"].(string)
			result[constants.FieldBlockDeviceProvisionerLonghorn] = []interface{}{
				map[string]interface{}{
					constants.FieldBlockDeviceProvisionerLonghornEV: engineVersion,
					constants.FieldBlockDeviceProvisionerLonghornDD: diskDriver,
				},
			}
		}
	}

	if lvm, ok := prov["lvm"]; ok {
		if lvmMap, ok := lvm.(map[string]interface{}); ok {
			vgName, _ := lvmMap["vgName"].(string)
			lvmResult := map[string]interface{}{
				constants.FieldBlockDeviceProvisionerLVMVGName: vgName,
			}
			if params, ok := lvmMap["parameters"]; ok {
				if paramSlice, ok := params.([]interface{}); ok {
					lvmResult[constants.FieldBlockDeviceProvisionerLVMParameters] = paramSlice
				} else {
					lvmResult[constants.FieldBlockDeviceProvisionerLVMParameters] = []interface{}{}
				}
			} else {
				lvmResult[constants.FieldBlockDeviceProvisionerLVMParameters] = []interface{}{}
			}
			result[constants.FieldBlockDeviceProvisionerLVM] = []interface{}{lvmResult}
		}
	}

	return []interface{}{result}
}

func readDeviceStatus(obj *unstructured.Unstructured) []interface{} {
	ds, found, _ := unstructured.NestedMap(obj.Object, "status", "deviceStatus")
	if !found {
		return []interface{}{}
	}

	statusDevPath, _ := ds["devPath"].(string)
	parentDevice, _ := ds["parentDevice"].(string)
	partitioned, _ := ds["partitioned"].(bool)

	var capacitySizeBytes int64
	if capacity, ok := ds["capacity"].(map[string]interface{}); ok {
		capacitySizeBytes = nestedInt64(capacity, "sizeBytes")
	}

	var deviceType, driveType, storageController, vendor, model, serialNumber, wwn, busPath string
	var isRemovable bool
	if details, ok := ds["details"].(map[string]interface{}); ok {
		deviceType, _ = details["deviceType"].(string)
		driveType, _ = details["driveType"].(string)
		storageController, _ = details["storageController"].(string)
		vendor, _ = details["vendor"].(string)
		model, _ = details["model"].(string)
		serialNumber, _ = details["serialNumber"].(string)
		wwn, _ = details["wwn"].(string)
		busPath, _ = details["busPath"].(string)
		isRemovable, _ = details["isRemovable"].(bool)
	}

	var fsType, mountPoint string
	var isReadOnly bool
	if fs, ok := ds["fileSystem"].(map[string]interface{}); ok {
		fsType, _ = fs["type"].(string)
		mountPoint, _ = fs["mountPoint"].(string)
		isReadOnly, _ = fs["isReadOnly"].(bool)
	}

	return []interface{}{
		map[string]interface{}{
			constants.FieldBlockDeviceStatusDevPath:           statusDevPath,
			constants.FieldBlockDeviceStatusParentDevice:      parentDevice,
			constants.FieldBlockDeviceStatusPartitioned:       partitioned,
			constants.FieldBlockDeviceStatusCapacitySizeBytes: int(capacitySizeBytes),
			constants.FieldBlockDeviceStatusDeviceType:        deviceType,
			constants.FieldBlockDeviceStatusDriveType:         driveType,
			constants.FieldBlockDeviceStatusStorageController: storageController,
			constants.FieldBlockDeviceStatusVendor:            vendor,
			constants.FieldBlockDeviceStatusModel:             model,
			constants.FieldBlockDeviceStatusSerialNumber:      serialNumber,
			constants.FieldBlockDeviceStatusWWN:               wwn,
			constants.FieldBlockDeviceStatusBusPath:           busPath,
			constants.FieldBlockDeviceStatusFSType:            fsType,
			constants.FieldBlockDeviceStatusMountPoint:        mountPoint,
			constants.FieldBlockDeviceStatusIsReadOnly:        isReadOnly,
			constants.FieldBlockDeviceStatusIsRemovable:       isRemovable,
		},
	}
}

// nestedInt64 extracts an int64 from a map, handling both int64 and float64 JSON types.
func nestedInt64(m map[string]interface{}, key string) int64 {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch val := v.(type) {
	case int64:
		return val
	case float64:
		return int64(val)
	default:
		return 0
	}
}

// getBlockDeviceLabels filters out NDM-managed and harvester-managed labels.
func getBlockDeviceLabels(labels map[string]string) map[string]string {
	filtered := map[string]string{}
	for key, value := range labels {
		if key == "kubernetes.io/hostname" ||
			strings.HasPrefix(key, "ndm.harvesterhci.io/") ||
			strings.HasPrefix(key, "tags.harvesterhci.io/") ||
			strings.HasPrefix(key, "harvesterhci.io/") {
			continue
		}
		filtered[key] = value
	}
	return filtered
}
