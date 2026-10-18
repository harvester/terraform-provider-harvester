package importer

import (
	"reflect"
	"testing"

	"github.com/harvester/harvester/pkg/builder"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

// Values shared by the kube-ovn iptables NAT importer tests.
const (
	testNatSubnetName      = "test-subnet"
	testNatGatewayName     = "test-gw"
	testNatEIPName         = "test-eip"
	testNatExternalSubnet  = "test-external"
	testNatExternalIP      = "192.168.1.100"
	testNatExternalIPv6    = "fd00:192:168::100"
	testNatInternalIP      = "10.0.0.100"
	testNatQoSPolicy       = "test-qos"
	testNatDescription     = "test description"
	testNatTagKey          = "env"
	testNatTagValue        = "test"
	testNatLabelKey        = "team"
	testNatLabelValue      = "net"
	testNatKubeOVNLabelKey = "ovn.kubernetes.io/vpc-nat-gw-name"
)

// testNatObjectMeta returns the metadata of a fully populated object: a
// description, a tag, a user label and a label set by the kube-ovn
// controller, which must not be read back.
func testNatObjectMeta(name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name: name,
		Labels: map[string]string{
			builder.LabelPrefixHarvesterTag + testNatTagKey: testNatTagValue,
			testNatLabelKey:        testNatLabelValue,
			testNatKubeOVNLabelKey: testNatGatewayName,
		},
		Annotations: map[string]string{
			builder.AnnotationKeyDescription: testNatDescription,
		},
	}
}

// testNatMetadataStates returns the states testNatObjectMeta maps to, or the
// states of an object without metadata when full is false.
func testNatMetadataStates(name string, full bool) map[string]interface{} {
	states := map[string]interface{}{
		constants.FieldCommonName:        name,
		constants.FieldCommonDescription: "",
		constants.FieldCommonTags:        map[string]string{},
		constants.FieldCommonLabels:      map[string]string{},
	}
	if full {
		states[constants.FieldCommonDescription] = testNatDescription
		states[constants.FieldCommonTags] = map[string]string{testNatTagKey: testNatTagValue}
		states[constants.FieldCommonLabels] = map[string]string{testNatLabelKey: testNatLabelValue}
	}
	return states
}

// checkNatStateGetter compares every state of a getter, its ID and its type.
func checkNatStateGetter(t *testing.T, getter *StateGetter, err error, name, resourceType string, expected map[string]interface{}) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if getter.ID != helper.BuildID("", name) || getter.Name != name || getter.ResourceType != resourceType {
		t.Errorf("ID, Name, ResourceType = %q, %q, %q, want %q, %q, %q",
			getter.ID, getter.Name, getter.ResourceType, helper.BuildID("", name), name, resourceType)
	}
	for key, want := range expected {
		if got, ok := getter.States[key]; !ok || !reflect.DeepEqual(got, want) {
			t.Errorf("%s = %#v, want %#v", key, got, want)
		}
	}
	for key := range getter.States {
		if _, ok := expected[key]; !ok {
			t.Errorf("unexpected state %s = %#v", key, getter.States[key])
		}
	}
}

// withStates returns states completed with extra.
func withStates(states, extra map[string]interface{}) map[string]interface{} {
	for key, value := range extra {
		states[key] = value
	}
	return states
}

func TestResourceKubeOVNVpcNatGatewayStateGetter(t *testing.T) {
	const (
		otherExternalSubnet = "test-external-2"
		selector            = "kubernetes.io/os: linux"
	)

	testcases := []struct {
		name     string
		gw       *kubeovnv1.VpcNatGateway
		expected map[string]interface{}
	}{
		{
			name: "gateway with all fields",
			gw: &kubeovnv1.VpcNatGateway{
				ObjectMeta: testNatObjectMeta(testNatGatewayName),
				Spec: kubeovnv1.VpcNatGatewaySpec{
					Vpc:             testVpcName,
					Subnet:          testNatSubnetName,
					LanIP:           testNatInternalIP,
					ExternalSubnets: []string{testNatExternalSubnet, otherExternalSubnet},
					Selector:        []string{selector},
					QoSPolicy:       testNatQoSPolicy,
				},
				Status: kubeovnv1.VpcNatGatewayStatus{
					QoSPolicy:       testNatQoSPolicy,
					ExternalSubnets: []string{testNatExternalSubnet},
					Selector:        []string{selector},
				},
			},
			expected: withStates(testNatMetadataStates(testNatGatewayName, true), map[string]interface{}{
				constants.FieldCommonState:                    constants.StateCommonActive,
				constants.FieldKubeOVNVpcNatGwVpc:             testVpcName,
				constants.FieldKubeOVNVpcNatGwSubnet:          testNatSubnetName,
				constants.FieldKubeOVNVpcNatGwLanIP:           testNatInternalIP,
				constants.FieldKubeOVNVpcNatGwExternalSubnets: []string{testNatExternalSubnet, otherExternalSubnet},
				constants.FieldKubeOVNVpcNatGwSelector:        []string{selector},
				constants.FieldKubeOVNVpcNatGwQoSPolicy:       testNatQoSPolicy,
				constants.FieldKubeOVNVpcNatGwStatusQoS:       testNatQoSPolicy,
				constants.FieldKubeOVNVpcNatGwStatusExtSubs:   []string{testNatExternalSubnet},
				constants.FieldKubeOVNVpcNatGwStatusSelector:  []string{selector},
			}),
		},
		{
			name: "empty gateway",
			gw:   &kubeovnv1.VpcNatGateway{ObjectMeta: metav1.ObjectMeta{Name: testNatGatewayName}},
			expected: withStates(testNatMetadataStates(testNatGatewayName, false), map[string]interface{}{
				constants.FieldCommonState:                    constants.StateCommonActive,
				constants.FieldKubeOVNVpcNatGwVpc:             "",
				constants.FieldKubeOVNVpcNatGwSubnet:          "",
				constants.FieldKubeOVNVpcNatGwLanIP:           "",
				constants.FieldKubeOVNVpcNatGwExternalSubnets: []string(nil),
				constants.FieldKubeOVNVpcNatGwSelector:        []string(nil),
				constants.FieldKubeOVNVpcNatGwQoSPolicy:       "",
				constants.FieldKubeOVNVpcNatGwStatusQoS:       "",
				constants.FieldKubeOVNVpcNatGwStatusExtSubs:   []string(nil),
				constants.FieldKubeOVNVpcNatGwStatusSelector:  []string(nil),
			}),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNVpcNatGatewayStateGetter(tc.gw)
			checkNatStateGetter(t, getter, err, testNatGatewayName, constants.ResourceTypeKubeOVNVpcNatGateway, tc.expected)
		})
	}
}
