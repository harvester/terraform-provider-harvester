package util

import (
	"testing"
)

func TestLabelsKeepsManagedLabels(t *testing.T) {
	labels := map[string]string{
		"tag.harvesterhci.io/role":          "db",
		"harvesterhci.io/creator":           "terraform-provider-harvester",
		"ovn.kubernetes.io/vpc-nat-gw-name": "gw",
		"ovn.kubernetes.io/eip_v4_ip":       "172.16.2.66",
		"stale-user-label":                  "x",
	}
	processors := NewProcessors().Labels(&labels)
	if err := processors[0].Parser(map[string]interface{}{"tier": "fast"}); err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"tag.harvesterhci.io/role":          "db",
		"harvesterhci.io/creator":           "terraform-provider-harvester",
		"ovn.kubernetes.io/vpc-nat-gw-name": "gw",
		"ovn.kubernetes.io/eip_v4_ip":       "172.16.2.66",
		"tier":                              "fast",
	}
	if len(labels) != len(want) {
		t.Fatalf("labels = %v, want %v", labels, want)
	}
	for k, v := range want {
		if labels[k] != v {
			t.Errorf("label %s = %q, want %q", k, labels[k], v)
		}
	}
}
