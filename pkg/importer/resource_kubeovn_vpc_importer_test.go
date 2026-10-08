package importer

import (
	"maps"
	"slices"
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

// Values shared by the kube-ovn importer tests.
const (
	testVpcName   = "test-vpc"
	testCIDR      = "10.0.0.0/24"
	testGatewayIP = "10.0.0.1"
)

// blockStrings returns the string field of every exported block, in order.
func blockStrings(blocks interface{}, field string) []string {
	list := blocks.([]map[string]interface{})
	values := make([]string, 0, len(list))
	for _, block := range list {
		values = append(values, block[field].(string))
	}
	return values
}

func TestResourceKubeOVNVpcStateGetter(t *testing.T) {
	testcases := []struct {
		name                 string
		vpc                  *kubeovnv1.Vpc
		expectedState        string
		expectedExternal     bool
		expectedBfd          bool
		expectedNamespaces   []string
		expectedStaticRoutes []string // CIDRs, in order
		expectedPolicyRoutes []string // matches, in order
		expectedDesc         string
		expectedTags         map[string]string
	}{
		{
			name: "vpc with static routes and namespaces",
			vpc: &kubeovnv1.Vpc{
				ObjectMeta: metav1.ObjectMeta{
					Name: testVpcName,
					Labels: map[string]string{
						"tag.harvesterhci.io/env": "test",
					},
					Annotations: map[string]string{
						"field.cattle.io/description": "Test VPC",
					},
				},
				Spec: kubeovnv1.VpcSpec{
					Namespaces:     []string{"ns1", "ns2"},
					EnableExternal: true,
					EnableBfd:      false,
					StaticRoutes: []*kubeovnv1.StaticRoute{
						{
							Policy:    "policySrc",
							CIDR:      testCIDR,
							NextHopIP: testGatewayIP,
						},
					},
				},
				Status: kubeovnv1.VpcStatus{
					Standby:              true,
					DefaultLogicalSwitch: "test-vpc-default",
					Router:               "test-vpc-router",
					Subnets:              []string{"subnet1"},
				},
			},
			expectedState:        constants.StateCommonActive,
			expectedExternal:     true,
			expectedNamespaces:   []string{"ns1", "ns2"},
			expectedStaticRoutes: []string{testCIDR},
			expectedDesc:         "Test VPC",
			expectedTags:         map[string]string{"env": "test"},
		},
		{
			name: "several static and policy routes keep their order",
			vpc: &kubeovnv1.Vpc{
				ObjectMeta: metav1.ObjectMeta{Name: "routed-vpc"},
				Spec: kubeovnv1.VpcSpec{
					EnableBfd: true,
					StaticRoutes: []*kubeovnv1.StaticRoute{
						{CIDR: "10.2.0.0/16", NextHopIP: "10.0.0.2"},
						{CIDR: "10.1.0.0/16", NextHopIP: testGatewayIP},
						{CIDR: "0.0.0.0/0", NextHopIP: "10.0.0.254"},
					},
					PolicyRoutes: []*kubeovnv1.PolicyRoute{
						{Priority: 200, Match: "ip4.src == 10.1.0.0/16", Action: kubeovnv1.PolicyRouteActionDrop},
						{Priority: 100, Match: "ip4.dst == 10.2.0.0/16", Action: kubeovnv1.PolicyRouteActionAllow},
					},
				},
			},
			expectedState:        constants.StateCommonActive,
			expectedBfd:          true,
			expectedStaticRoutes: []string{"10.2.0.0/16", "10.1.0.0/16", "0.0.0.0/0"},
			expectedPolicyRoutes: []string{"ip4.src == 10.1.0.0/16", "ip4.dst == 10.2.0.0/16"},
		},
		{
			name: "ready condition marks the vpc ready",
			vpc: &kubeovnv1.Vpc{
				ObjectMeta: metav1.ObjectMeta{Name: "ready-vpc"},
				Status: kubeovnv1.VpcStatus{Conditions: []kubeovnv1.Condition{
					{Type: kubeovnv1.Validated, Status: corev1.ConditionTrue},
					{Type: kubeovnv1.Ready, Status: corev1.ConditionTrue},
				}},
			},
			expectedState: constants.StateCommonReady,
		},
		{
			name: "ready condition not true keeps the vpc active",
			vpc: &kubeovnv1.Vpc{
				ObjectMeta: metav1.ObjectMeta{Name: "pending-vpc"},
				Status: kubeovnv1.VpcStatus{Conditions: []kubeovnv1.Condition{
					{Type: kubeovnv1.Validated, Status: corev1.ConditionTrue},
					{Type: kubeovnv1.Ready, Status: corev1.ConditionFalse},
				}},
			},
			expectedState: constants.StateCommonActive,
		},
		{
			name: "empty vpc with nil labels",
			vpc: &kubeovnv1.Vpc{
				ObjectMeta: metav1.ObjectMeta{
					Name: "empty-vpc",
				},
				Spec: kubeovnv1.VpcSpec{},
			},
			expectedState: constants.StateCommonActive,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNVpcStateGetter(tc.vpc)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if getter.ID != helper.BuildID("", tc.vpc.Name) {
				t.Errorf("ID: expected %q, got %q", helper.BuildID("", tc.vpc.Name), getter.ID)
			}
			if getter.Name != tc.vpc.Name {
				t.Errorf("Name: expected %q, got %q", tc.vpc.Name, getter.Name)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNVpc {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeKubeOVNVpc, getter.ResourceType)
			}

			state := getter.States[constants.FieldCommonState].(string)
			if state != tc.expectedState {
				t.Errorf("State: expected %q, got %q", tc.expectedState, state)
			}

			enableExternal := getter.States[constants.FieldKubeOVNVpcEnableExternal].(bool)
			if enableExternal != tc.expectedExternal {
				t.Errorf("EnableExternal: expected %v, got %v", tc.expectedExternal, enableExternal)
			}

			enableBfd := getter.States[constants.FieldKubeOVNVpcEnableBfd].(bool)
			if enableBfd != tc.expectedBfd {
				t.Errorf("EnableBfd: expected %v, got %v", tc.expectedBfd, enableBfd)
			}

			namespaces := getter.States[constants.FieldKubeOVNVpcNamespaces].([]string)
			if !slices.Equal(namespaces, tc.expectedNamespaces) {
				t.Errorf("Namespaces: expected %v, got %v", tc.expectedNamespaces, namespaces)
			}

			staticRoutes := blockStrings(getter.States[constants.FieldKubeOVNVpcStaticRoutes], constants.FieldKubeOVNStaticRouteCIDR)
			if !slices.Equal(staticRoutes, tc.expectedStaticRoutes) {
				t.Errorf("StaticRoutes: expected %v, got %v", tc.expectedStaticRoutes, staticRoutes)
			}

			policyRoutes := blockStrings(getter.States[constants.FieldKubeOVNVpcPolicyRoutes], constants.FieldKubeOVNPolicyRouteMatch)
			if !slices.Equal(policyRoutes, tc.expectedPolicyRoutes) {
				t.Errorf("PolicyRoutes: expected %v, got %v", tc.expectedPolicyRoutes, policyRoutes)
			}

			desc := getter.States[constants.FieldCommonDescription].(string)
			if desc != tc.expectedDesc {
				t.Errorf("Description: expected %q, got %q", tc.expectedDesc, desc)
			}

			tags := getter.States[constants.FieldCommonTags].(map[string]string)
			if !maps.Equal(tags, tc.expectedTags) {
				t.Errorf("Tags: expected %v, got %v", tc.expectedTags, tags)
			}
		})
	}
}
