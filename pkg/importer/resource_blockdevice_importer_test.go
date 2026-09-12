package importer

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func newFullBlockDevice() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "harvesterhci.io/v1beta1",
			"kind":       "BlockDevice",
			"metadata": map[string]interface{}{
				"name":      "abc123def456",
				"namespace": "longhorn-system",
				"labels": map[string]interface{}{
					"kubernetes.io/hostname":          "node1",
					"ndm.harvesterhci.io/device-type": "disk",
				},
				"annotations": map[string]interface{}{
					"field.cattle.io/description": "test block device",
				},
			},
			"spec": map[string]interface{}{
				"devPath":   "/dev/sda",
				"nodeName":  "node1",
				"provision": true,
				"fileSystem": map[string]interface{}{
					"mountPoint":     "",
					"forceFormatted": false,
				},
				"tags": []interface{}{"default", "ssd"},
				"provisioner": map[string]interface{}{
					"longhorn": map[string]interface{}{
						"engineVersion": "LonghornV2",
						"diskDriver":    "auto",
					},
				},
			},
			"status": map[string]interface{}{
				"provisionPhase": "Provisioned",
				"state":          "Active",
				"deviceStatus": map[string]interface{}{
					"devPath":     "/dev/sda",
					"partitioned": true,
					"capacity": map[string]interface{}{
						"sizeBytes":              float64(500107862016),
						"physicalBlockSizeBytes": float64(512),
					},
					"details": map[string]interface{}{
						"deviceType":        "disk",
						"driveType":         "SSD",
						"storageController": "NVMe",
						"vendor":            "Samsung",
						"model":             "970 EVO",
						"serialNumber":      "S123456789",
						"wwn":               "naa.12345",
						"busPath":           "pci-0000:00:1f.2",
						"isRemovable":       false,
					},
					"fileSystem": map[string]interface{}{
						"type":       "ext4",
						"mountPoint": "/var/lib/longhorn",
						"isReadOnly": false,
					},
				},
			},
		},
	}
}

func TestBlockDeviceStateGetterIdentity(t *testing.T) {
	sg, err := ResourceBlockDeviceStateGetter(newFullBlockDevice())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sg.ID != "longhorn-system/abc123def456" {
		t.Errorf("expected ID longhorn-system/abc123def456, got %s", sg.ID)
	}
	if sg.Name != "abc123def456" {
		t.Errorf("expected Name abc123def456, got %s", sg.Name)
	}
	if sg.ResourceType != "harvester_blockdevice" {
		t.Errorf("expected ResourceType harvester_blockdevice, got %s", sg.ResourceType)
	}
}

func TestBlockDeviceStateGetterSpec(t *testing.T) {
	sg, err := ResourceBlockDeviceStateGetter(newFullBlockDevice())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sg.States["node_name"] != "node1" {
		t.Errorf("expected node_name=node1, got %v", sg.States["node_name"])
	}
	if sg.States["dev_path"] != "/dev/sda" {
		t.Errorf("expected dev_path=/dev/sda, got %v", sg.States["dev_path"])
	}
	if sg.States["provision"] != true {
		t.Errorf("expected provision=true, got %v", sg.States["provision"])
	}
	if sg.States["force_formatted"] != false {
		t.Errorf("expected force_formatted=false, got %v", sg.States["force_formatted"])
	}

	deviceTags := sg.States["device_tags"].([]interface{})
	if len(deviceTags) != 2 {
		t.Errorf("expected 2 device tags, got %d", len(deviceTags))
	}
}

func TestBlockDeviceStateGetterProvisioner(t *testing.T) {
	sg, err := ResourceBlockDeviceStateGetter(newFullBlockDevice())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prov := sg.States["disk_provisioner"].([]interface{})
	if len(prov) != 1 {
		t.Fatalf("expected 1 provisioner block, got %d", len(prov))
	}
	provMap := prov[0].(map[string]interface{})
	lhList := provMap["longhorn"].([]interface{})
	if len(lhList) != 1 {
		t.Fatalf("expected 1 longhorn block, got %d", len(lhList))
	}
	lh := lhList[0].(map[string]interface{})
	if lh["engine_version"] != "LonghornV2" {
		t.Errorf("expected engine_version=LonghornV2, got %v", lh["engine_version"])
	}
}

func TestBlockDeviceStateGetterStatus(t *testing.T) {
	sg, err := ResourceBlockDeviceStateGetter(newFullBlockDevice())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sg.States["provision_phase"] != "Provisioned" {
		t.Errorf("expected provision_phase=Provisioned, got %v", sg.States["provision_phase"])
	}
	if sg.States["state"] != "Active" {
		t.Errorf("expected state=Active, got %v", sg.States["state"])
	}

	dsList := sg.States["device_status"].([]interface{})
	if len(dsList) != 1 {
		t.Fatalf("expected 1 device_status block, got %d", len(dsList))
	}
	ds := dsList[0].(map[string]interface{})
	if ds["capacity_size_bytes"] != int(500107862016) {
		t.Errorf("expected capacity_size_bytes=500107862016, got %v", ds["capacity_size_bytes"])
	}
	if ds["vendor"] != "Samsung" {
		t.Errorf("expected vendor=Samsung, got %v", ds["vendor"])
	}
	if ds["device_type"] != "disk" {
		t.Errorf("expected device_type=disk, got %v", ds["device_type"])
	}
}

func TestBlockDeviceStateGetterLabels(t *testing.T) {
	sg, err := ResourceBlockDeviceStateGetter(newFullBlockDevice())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	labels := sg.States["labels"].(map[string]string)
	if _, ok := labels["kubernetes.io/hostname"]; ok {
		t.Error("kubernetes.io/hostname should be filtered from labels")
	}
	if _, ok := labels["ndm.harvesterhci.io/device-type"]; ok {
		t.Error("ndm.harvesterhci.io/device-type should be filtered from labels")
	}
}

func TestBlockDeviceStateGetterMinimal(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "harvesterhci.io/v1beta1",
			"kind":       "BlockDevice",
			"metadata": map[string]interface{}{
				"name":      "min-device",
				"namespace": "longhorn-system",
			},
			"spec": map[string]interface{}{
				"devPath":   "/dev/sdb",
				"nodeName":  "node2",
				"provision": false,
				"fileSystem": map[string]interface{}{
					"mountPoint": "",
				},
			},
			"status": map[string]interface{}{
				"provisionPhase": "Unprovisioned",
				"state":          "Inactive",
			},
		},
	}

	sg, err := ResourceBlockDeviceStateGetter(obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sg.ID != "longhorn-system/min-device" {
		t.Errorf("expected ID longhorn-system/min-device, got %s", sg.ID)
	}
	if sg.States["provision"] != false {
		t.Errorf("expected provision=false, got %v", sg.States["provision"])
	}
	if sg.States["provision_phase"] != "Unprovisioned" {
		t.Errorf("expected provision_phase=Unprovisioned, got %v", sg.States["provision_phase"])
	}

	prov := sg.States["disk_provisioner"].([]interface{})
	if len(prov) != 0 {
		t.Errorf("expected empty provisioner, got %v", prov)
	}

	ds := sg.States["device_status"].([]interface{})
	if len(ds) != 0 {
		t.Errorf("expected empty device_status, got %v", ds)
	}
}
