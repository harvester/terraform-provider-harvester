package kubeovn_vpc_nat_gateway

import (
	"context"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/util"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func TestVpcNatGatewayConstructor(t *testing.T) {
	const (
		externalSubnet1 = "ext-sub-1"
		externalSubnet2 = "ext-sub-2"
		linuxSelector   = "kubernetes.io/os=linux"
		nodeSelector    = "kubernetes.io/hostname=node1"
		qosPolicy       = "qos-1"
	)

	existing := &kubeovnv1.VpcNatGateway{
		ObjectMeta: metav1.ObjectMeta{Name: "test-gw"},
		Spec: kubeovnv1.VpcNatGatewaySpec{
			ExternalSubnets: []string{"old-ext-sub"},
			Selector:        []string{"old=selector"},
			QoSPolicy:       "old-qos",
		},
	}

	testcases := []struct {
		name                    string
		existing                *kubeovnv1.VpcNatGateway // nil creates the gateway
		config                  map[string]interface{}
		expectedExternalSubnets []string
		expectedSelector        []string
		expectedQoSPolicy       string
	}{
		{
			name:   "create without lists",
			config: map[string]interface{}{},
		},
		{
			name: "create keeps the order of several entries",
			config: map[string]interface{}{
				constants.FieldKubeOVNVpcNatGwExternalSubnets: []interface{}{externalSubnet2, externalSubnet1, "ext-sub-3"},
				constants.FieldKubeOVNVpcNatGwSelector:        []interface{}{nodeSelector, linuxSelector},
				constants.FieldKubeOVNVpcNatGwQoSPolicy:       qosPolicy,
			},
			expectedExternalSubnets: []string{externalSubnet2, externalSubnet1, "ext-sub-3"},
			expectedSelector:        []string{nodeSelector, linuxSelector},
			expectedQoSPolicy:       qosPolicy,
		},
		{
			name:     "update drops lists and the QoS policy removed from the configuration",
			existing: existing,
			config:   map[string]interface{}{},
		},
		{
			name:     "update replaces lists instead of appending to them",
			existing: existing,
			config: map[string]interface{}{
				constants.FieldKubeOVNVpcNatGwExternalSubnets: []interface{}{externalSubnet1, externalSubnet2},
				constants.FieldKubeOVNVpcNatGwSelector:        []interface{}{linuxSelector, nodeSelector},
				constants.FieldKubeOVNVpcNatGwQoSPolicy:       qosPolicy,
			},
			expectedExternalSubnets: []string{externalSubnet1, externalSubnet2},
			expectedSelector:        []string{linuxSelector, nodeSelector},
			expectedQoSPolicy:       qosPolicy,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			constructor := Creator("test-gw")
			if tc.existing != nil {
				constructor = Updater(tc.existing.DeepCopy())
			}
			d := schema.TestResourceDataRaw(t, Schema(), tc.config)
			result, err := util.ResourceConstruct(context.Background(), d, constructor)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			gateway := result.(*kubeovnv1.VpcNatGateway)

			if !slices.Equal(gateway.Spec.ExternalSubnets, tc.expectedExternalSubnets) {
				t.Errorf("ExternalSubnets: expected %v, got %v", tc.expectedExternalSubnets, gateway.Spec.ExternalSubnets)
			}
			if !slices.Equal(gateway.Spec.Selector, tc.expectedSelector) {
				t.Errorf("Selector: expected %v, got %v", tc.expectedSelector, gateway.Spec.Selector)
			}
			if gateway.Spec.QoSPolicy != tc.expectedQoSPolicy {
				t.Errorf("QoSPolicy: expected %q, got %q", tc.expectedQoSPolicy, gateway.Spec.QoSPolicy)
			}
		})
	}
}
