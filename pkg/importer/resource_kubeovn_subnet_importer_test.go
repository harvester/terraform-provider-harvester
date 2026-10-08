package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func boolPtr(b bool) *bool {
	return &b
}

func TestResourceKubeOVNSubnetStateGetter(t *testing.T) {
	testcases := []struct {
		name             string
		subnet           *kubeovnv1.Subnet
		expectedID       string
		expectedState    string
		expectedVpc      string
		expectedCIDR     string
		expectedGateway  string
		expectedProtocol string
		expectedEnableLb bool
		expectedDesc     string
		expectedTags     map[string]string
	}{
		{
			name: "subnet with all fields",
			subnet: &kubeovnv1.Subnet{
				ObjectMeta: metav1.ObjectMeta{
					Name: testSubnetName,
					Labels: map[string]string{
						"tag.harvesterhci.io/env": "production",
					},
					Annotations: map[string]string{
						"field.cattle.io/description": "Test subnet",
					},
				},
				Spec: kubeovnv1.SubnetSpec{
					Vpc:         testVpcName,
					CIDRBlock:   testCIDR,
					Gateway:     testGatewayIP,
					Protocol:    "IPv4",
					ExcludeIps:  []string{testGatewayIP},
					NatOutgoing: true,
					GatewayType: "distributed",
					EnableLb:    boolPtr(true),
				},
				Status: kubeovnv1.SubnetStatus{
					V4AvailableIPs: 253,
					V4UsingIPs:     1,
				},
			},
			expectedID:       helper.BuildID("", testSubnetName),
			expectedState:    constants.StateCommonActive,
			expectedVpc:      testVpcName,
			expectedCIDR:     testCIDR,
			expectedGateway:  testGatewayIP,
			expectedProtocol: "IPv4",
			expectedEnableLb: true,
			expectedDesc:     "Test subnet",
			expectedTags:     map[string]string{"env": "production"},
		},
		{
			name: "subnet with nil EnableLb defaults to true",
			subnet: &kubeovnv1.Subnet{
				ObjectMeta: metav1.ObjectMeta{
					Name: "minimal-subnet",
				},
				Spec: kubeovnv1.SubnetSpec{
					CIDRBlock: "192.168.0.0/16",
					Gateway:   "192.168.0.1",
				},
			},
			expectedID:       helper.BuildID("", "minimal-subnet"),
			expectedState:    constants.StateCommonActive,
			expectedVpc:      "",
			expectedCIDR:     "192.168.0.0/16",
			expectedGateway:  "192.168.0.1",
			expectedProtocol: "",
			expectedEnableLb: true,
			expectedDesc:     "",
			expectedTags:     map[string]string{},
		},
		{
			name: "subnet with EnableLb false",
			subnet: &kubeovnv1.Subnet{
				ObjectMeta: metav1.ObjectMeta{
					Name: "no-lb-subnet",
				},
				Spec: kubeovnv1.SubnetSpec{
					CIDRBlock: "172.16.0.0/16",
					Gateway:   "172.16.0.1",
					EnableLb:  boolPtr(false),
				},
			},
			expectedID:       helper.BuildID("", "no-lb-subnet"),
			expectedState:    constants.StateCommonActive,
			expectedVpc:      "",
			expectedCIDR:     "172.16.0.0/16",
			expectedGateway:  "172.16.0.1",
			expectedProtocol: "",
			expectedEnableLb: false,
			expectedDesc:     "",
			expectedTags:     map[string]string{},
		},
		{
			name: "ready condition marks the subnet ready",
			subnet: &kubeovnv1.Subnet{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ready-subnet",
				},
				Status: kubeovnv1.SubnetStatus{Conditions: []kubeovnv1.Condition{
					{Type: kubeovnv1.Validated, Status: corev1.ConditionTrue},
					{Type: kubeovnv1.Ready, Status: corev1.ConditionTrue},
				}},
			},
			expectedID:       helper.BuildID("", "ready-subnet"),
			expectedState:    constants.StateCommonReady,
			expectedEnableLb: true,
			expectedTags:     map[string]string{},
		},
		{
			name: "ready condition not true keeps the subnet active",
			subnet: &kubeovnv1.Subnet{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pending-subnet",
				},
				Status: kubeovnv1.SubnetStatus{Conditions: []kubeovnv1.Condition{
					{Type: kubeovnv1.Validated, Status: corev1.ConditionTrue},
					{Type: kubeovnv1.Ready, Status: corev1.ConditionUnknown},
				}},
			},
			expectedID:       helper.BuildID("", "pending-subnet"),
			expectedState:    constants.StateCommonActive,
			expectedEnableLb: true,
			expectedTags:     map[string]string{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNSubnetStateGetter(tc.subnet)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.Name != tc.subnet.Name {
				t.Errorf("Name: expected %q, got %q", tc.subnet.Name, getter.Name)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNSubnet {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeKubeOVNSubnet, getter.ResourceType)
			}

			state := getter.States[constants.FieldCommonState].(string)
			if state != tc.expectedState {
				t.Errorf("State: expected %q, got %q", tc.expectedState, state)
			}

			vpc := getter.States[constants.FieldKubeOVNSubnetVpc].(string)
			if vpc != tc.expectedVpc {
				t.Errorf("Vpc: expected %q, got %q", tc.expectedVpc, vpc)
			}

			cidr := getter.States[constants.FieldKubeOVNSubnetCIDRBlock].(string)
			if cidr != tc.expectedCIDR {
				t.Errorf("CIDRBlock: expected %q, got %q", tc.expectedCIDR, cidr)
			}

			gateway := getter.States[constants.FieldKubeOVNSubnetGateway].(string)
			if gateway != tc.expectedGateway {
				t.Errorf("Gateway: expected %q, got %q", tc.expectedGateway, gateway)
			}

			enableLb := getter.States[constants.FieldKubeOVNSubnetEnableLb].(bool)
			if enableLb != tc.expectedEnableLb {
				t.Errorf("EnableLb: expected %v, got %v", tc.expectedEnableLb, enableLb)
			}

			desc := getter.States[constants.FieldCommonDescription].(string)
			if desc != tc.expectedDesc {
				t.Errorf("Description: expected %q, got %q", tc.expectedDesc, desc)
			}

			tags := getter.States[constants.FieldCommonTags].(map[string]string)
			if len(tags) != len(tc.expectedTags) {
				t.Errorf("Tags: expected %d, got %d: %v", len(tc.expectedTags), len(tags), tags)
			}
		})
	}
}
