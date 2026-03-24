package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceKubeOVNVpcStateGetter(t *testing.T) {
	testcases := []struct {
		name               string
		vpc                *kubeovnv1.Vpc
		expectedID         string
		expectedState      string
		expectedExternal   bool
		expectedBfd        bool
		expectedNamespaces int
		expectedRoutes     int
		expectedDesc       string
		expectedTags       map[string]string
	}{
		{
			name: "vpc with static routes and namespaces",
			vpc: &kubeovnv1.Vpc{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-vpc",
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
							CIDR:      "10.0.0.0/24",
							NextHopIP: "10.0.0.1",
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
			expectedID:         helper.BuildID("", "test-vpc"),
			expectedState:      constants.StateCommonActive,
			expectedExternal:   true,
			expectedBfd:        false,
			expectedNamespaces: 2,
			expectedRoutes:     1,
			expectedDesc:       "Test VPC",
			expectedTags:       map[string]string{"env": "test"},
		},
		{
			name: "empty vpc with nil labels",
			vpc: &kubeovnv1.Vpc{
				ObjectMeta: metav1.ObjectMeta{
					Name: "empty-vpc",
				},
				Spec: kubeovnv1.VpcSpec{},
			},
			expectedID:         helper.BuildID("", "empty-vpc"),
			expectedState:      constants.StateCommonActive,
			expectedExternal:   false,
			expectedBfd:        false,
			expectedNamespaces: 0,
			expectedRoutes:     0,
			expectedDesc:       "",
			expectedTags:       map[string]string{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNVpcStateGetter(tc.vpc)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
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
			if len(namespaces) != tc.expectedNamespaces {
				t.Errorf("Namespaces: expected %d, got %d", tc.expectedNamespaces, len(namespaces))
			}

			staticRoutes := getter.States[constants.FieldKubeOVNVpcStaticRoutes].([]map[string]interface{})
			if len(staticRoutes) != tc.expectedRoutes {
				t.Errorf("StaticRoutes: expected %d, got %d", tc.expectedRoutes, len(staticRoutes))
			}

			desc := getter.States[constants.FieldCommonDescription].(string)
			if desc != tc.expectedDesc {
				t.Errorf("Description: expected %q, got %q", tc.expectedDesc, desc)
			}

			tags := getter.States[constants.FieldCommonTags].(map[string]string)
			if len(tags) != len(tc.expectedTags) {
				t.Errorf("Tags: expected %d, got %d: %v", len(tc.expectedTags), len(tags), tags)
			}
			for key, val := range tc.expectedTags {
				if tags[key] != val {
					t.Errorf("Tag %q: expected %q, got %q", key, val, tags[key])
				}
			}
		})
	}
}
