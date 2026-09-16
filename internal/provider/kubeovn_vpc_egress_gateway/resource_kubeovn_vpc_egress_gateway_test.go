package kubeovn_vpc_egress_gateway

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// kube-ovn allocates the internal/external IPs (and fills the internal subnet
// from the VPC default) when the user leaves them unset, then writes them back
// into the spec. With those fields Optional+Computed, a configuration that
// omits them must plan no change once the values have been read back.
func TestVpcEgressGatewayComputedFieldsNoDrift(t *testing.T) {
	config := map[string]interface{}{
		constants.FieldCommonNamespace:                       "default",
		constants.FieldCommonName:                            "egw",
		constants.FieldKubeOVNVpcEgressGatewayExternalSubnet: "egress-external",
	}

	state := &terraform.InstanceState{
		ID: "default/egw",
		Attributes: map[string]string{
			"id":                           "default/egw",
			constants.FieldCommonNamespace: "default",
			constants.FieldCommonName:      "egw",
			constants.FieldKubeOVNVpcEgressGatewayExternalSubnet: "egress-external",
			// values kube-ovn populated and the importer read back
			constants.FieldKubeOVNVpcEgressGatewayInternalSubnet:     "ovn-default",
			constants.FieldKubeOVNVpcEgressGatewayInternalIPs + ".#": "1",
			constants.FieldKubeOVNVpcEgressGatewayInternalIPs + ".0": "10.54.0.5",
			constants.FieldKubeOVNVpcEgressGatewayExternalIPs + ".#": "1",
			constants.FieldKubeOVNVpcEgressGatewayExternalIPs + ".0": "172.16.3.204",
		},
	}

	diff, err := ResourceKubeOVNVpcEgressGateway().Diff(context.Background(), state, terraform.NewResourceConfigRaw(config), nil)
	if err != nil {
		t.Fatalf("unexpected diff error: %v", err)
	}
	if diff == nil || diff.Empty() {
		return
	}
	for key, attr := range diff.Attributes {
		for _, field := range []string{
			constants.FieldKubeOVNVpcEgressGatewayInternalSubnet,
			constants.FieldKubeOVNVpcEgressGatewayInternalIPs,
			constants.FieldKubeOVNVpcEgressGatewayExternalIPs,
		} {
			if strings.HasPrefix(key, field) {
				t.Errorf("unexpected drift on server-populated field %s: %+v", key, attr)
			}
		}
	}
}
