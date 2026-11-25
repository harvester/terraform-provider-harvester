package network

import (
	"context"
	"reflect"
	"testing"

	nadv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// TestUpdaterKeepsControllerLabels checks that an update of a network leaves
// the labels the Harvester network controller derives from the
// NetworkAttachmentDefinition on the object, together with the labels written
// by anyone else, so it does not revert them behind their back. The only label
// removed is the one the state carried and the configuration dropped.
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
				"example.com/owner":                      "someone-else",
				"environment":                            "dev",
			},
		},
	}
	// The tag is rebuilt from the tags of the configuration, so it is dropped
	// when the processors are registered.
	afterSetup := map[string]string{
		"network.harvesterhci.io/clusternetwork": "public",
		"network.harvesterhci.io/ready":          "true",
		"network.harvesterhci.io/type":           "L2VlanNetwork",
		"network.harvesterhci.io/vlan-id":        "2166",
		"example.com/owner":                      "someone-else",
		"environment":                            "dev",
	}
	expected := map[string]string{
		"network.harvesterhci.io/clusternetwork": "public",
		"network.harvesterhci.io/ready":          "true",
		"network.harvesterhci.io/type":           "L2VlanNetwork",
		"network.harvesterhci.io/vlan-id":        "2166",
		"example.com/owner":                      "someone-else",
	}

	constructor := Updater(nil, context.Background(), network)
	processors := constructor.Setup()

	if labels := network.Labels; !reflect.DeepEqual(labels, afterSetup) {
		t.Errorf("labels after the processors are registered = %v, want %v", labels, afterSetup)
	}

	var parser func(interface{}, interface{}) error
	for _, processor := range processors {
		if processor.Field == constants.FieldCommonLabels {
			parser = processor.ChangeParser
		}
	}
	if parser == nil {
		t.Fatalf("no processor registered for %s", constants.FieldCommonLabels)
	}

	// The state carries the label the user manages, the configuration no
	// longer declares it.
	state := map[string]interface{}{"environment": "dev"}
	configured := map[string]interface{}{}
	if err := parser(state, configured); err != nil {
		t.Fatalf("parsing the configured labels: %v", err)
	}

	if labels := network.Labels; !reflect.DeepEqual(labels, expected) {
		t.Errorf("labels after the processors run = %v, want %v", labels, expected)
	}
}
