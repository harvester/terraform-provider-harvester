package network

import (
	"context"
	"reflect"
	"testing"

	nadv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestUpdaterKeepsControllerLabels checks that the processors leave the labels
// the Harvester network controller derives from the NetworkAttachmentDefinition
// on the object, so an update does not remove them behind the controller's
// back. The labels the user manages are dropped here and added back from the
// configuration.
func TestUpdaterKeepsControllerLabels(t *testing.T) {
	network := &nadv1.NetworkAttachmentDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "public-vlan-2166",
			Namespace:   "harvester-public",
			Annotations: map[string]string{},
			Labels: map[string]string{
				"network.harvesterhci.io/clusternetwork": "public",
				"network.harvesterhci.io/ready":          "true",
				"network.harvesterhci.io/type":           "L2VlanNetwork",
				"network.harvesterhci.io/vlan-id":        "2166",
				"tag.harvesterhci.io/environment":        "dev",
				"environment":                            "dev",
			},
		},
	}
	expected := map[string]string{
		"network.harvesterhci.io/clusternetwork": "public",
		"network.harvesterhci.io/ready":          "true",
		"network.harvesterhci.io/type":           "L2VlanNetwork",
		"network.harvesterhci.io/vlan-id":        "2166",
	}

	constructor := Updater(nil, context.Background(), network)
	constructor.Setup()

	if labels := network.Labels; !reflect.DeepEqual(labels, expected) {
		t.Errorf("labels before the processors run = %v, want %v", labels, expected)
	}
}
