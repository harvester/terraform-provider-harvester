package importer

import "testing"

func TestGetLabelsSkipsManagedLabels(t *testing.T) {
	got := GetLabels(map[string]string{
		"tag.harvesterhci.io/role":          "db",
		"harvesterhci.io/creator":           "terraform-provider-harvester",
		"ovn.kubernetes.io/vpc-nat-gw-name": "gw",
		"ovn.kubernetes.io/eip_v4_ip":       "172.16.2.66",
		"tier":                              "fast",
	})
	if len(got) != 1 || got["tier"] != "fast" {
		t.Errorf("GetLabels() = %v, want only tier=fast", got)
	}
}
