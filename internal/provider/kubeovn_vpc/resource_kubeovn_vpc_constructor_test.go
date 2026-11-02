package kubeovn_vpc

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

func staticRoute(cidr string) map[string]interface{} {
	return map[string]interface{}{
		constants.FieldKubeOVNStaticRouteCIDR:      cidr,
		constants.FieldKubeOVNStaticRouteNextHopIP: "10.0.0.254",
	}
}

func policyRoute(priority int, match string) map[string]interface{} {
	return map[string]interface{}{
		constants.FieldKubeOVNPolicyRoutePriority: priority,
		constants.FieldKubeOVNPolicyRouteMatch:    match,
		constants.FieldKubeOVNPolicyRouteAction:   "drop",
	}
}

func TestVpcConstructor(t *testing.T) {
	const (
		namespace1 = "ns1"
		namespace2 = "ns2"
		cidr1      = "10.1.0.0/16"
		cidr2      = "10.2.0.0/16"
	)

	existing := &kubeovnv1.Vpc{
		ObjectMeta: metav1.ObjectMeta{Name: "test-vpc"},
		Spec: kubeovnv1.VpcSpec{
			Namespaces:   []string{"old-ns"},
			StaticRoutes: []*kubeovnv1.StaticRoute{{CIDR: "192.168.0.0/16", NextHopIP: "10.0.0.1"}},
			PolicyRoutes: []*kubeovnv1.PolicyRoute{{Priority: 1, Match: "ip4", Action: kubeovnv1.PolicyRouteActionAllow}},
		},
	}

	testcases := []struct {
		name                 string
		existing             *kubeovnv1.Vpc // nil creates the vpc
		config               map[string]interface{}
		expectedNamespaces   []string
		expectedStaticRoutes []string // CIDRs, in order
		expectedPolicyRoutes []string // matches, in order
	}{
		{
			name:   "create without lists",
			config: map[string]interface{}{},
		},
		{
			name: "create keeps the order of several entries",
			config: map[string]interface{}{
				constants.FieldKubeOVNVpcNamespaces: []interface{}{namespace2, namespace1, "ns3"},
				constants.FieldKubeOVNVpcStaticRoutes: []interface{}{
					staticRoute(cidr2), staticRoute(cidr1), staticRoute("0.0.0.0/0"),
				},
				constants.FieldKubeOVNVpcPolicyRoutes: []interface{}{
					policyRoute(200, "ip4.src == 10.1.0.0/16"), policyRoute(100, "ip4.dst == 10.2.0.0/16"),
				},
			},
			expectedNamespaces:   []string{namespace2, namespace1, "ns3"},
			expectedStaticRoutes: []string{cidr2, cidr1, "0.0.0.0/0"},
			expectedPolicyRoutes: []string{"ip4.src == 10.1.0.0/16", "ip4.dst == 10.2.0.0/16"},
		},
		{
			name:     "update drops lists removed from the configuration",
			existing: existing,
			config:   map[string]interface{}{},
		},
		{
			name:     "update replaces lists instead of appending to them",
			existing: existing,
			config: map[string]interface{}{
				constants.FieldKubeOVNVpcNamespaces:   []interface{}{namespace1, namespace2},
				constants.FieldKubeOVNVpcStaticRoutes: []interface{}{staticRoute(cidr1), staticRoute(cidr2)},
				constants.FieldKubeOVNVpcPolicyRoutes: []interface{}{policyRoute(100, "ip4"), policyRoute(50, "ip6")},
			},
			expectedNamespaces:   []string{namespace1, namespace2},
			expectedStaticRoutes: []string{cidr1, cidr2},
			expectedPolicyRoutes: []string{"ip4", "ip6"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			constructor := Creator("test-vpc")
			if tc.existing != nil {
				constructor = Updater(tc.existing.DeepCopy())
			}
			d := schema.TestResourceDataRaw(t, Schema(), tc.config)
			result, err := util.ResourceConstruct(context.Background(), d, constructor)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			vpc := result.(*kubeovnv1.Vpc)

			if !slices.Equal(vpc.Spec.Namespaces, tc.expectedNamespaces) {
				t.Errorf("Namespaces: expected %v, got %v", tc.expectedNamespaces, vpc.Spec.Namespaces)
			}
			var staticRoutes []string
			for _, route := range vpc.Spec.StaticRoutes {
				staticRoutes = append(staticRoutes, route.CIDR)
			}
			if !slices.Equal(staticRoutes, tc.expectedStaticRoutes) {
				t.Errorf("StaticRoutes: expected %v, got %v", tc.expectedStaticRoutes, staticRoutes)
			}
			var policyRoutes []string
			for _, route := range vpc.Spec.PolicyRoutes {
				policyRoutes = append(policyRoutes, route.Match)
			}
			if !slices.Equal(policyRoutes, tc.expectedPolicyRoutes) {
				t.Errorf("PolicyRoutes: expected %v, got %v", tc.expectedPolicyRoutes, policyRoutes)
			}
		})
	}
}

// TestVpcConstructorBooleans checks that every boolean is sent as configured,
// false included, so that an update can switch it off.
func TestVpcConstructorBooleans(t *testing.T) {
	const name = "test-vpc"
	testcases := []struct {
		name     string
		existing *kubeovnv1.Vpc // nil creates the VPC
		value    bool
	}{
		{name: "create with true", value: true},
		{name: "create with false", value: false},
		{name: "update switches every boolean off", existing: &kubeovnv1.Vpc{ObjectMeta: metav1.ObjectMeta{Name: name}, Spec: kubeovnv1.VpcSpec{EnableExternal: true, EnableBfd: true}}, value: false},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			d := schema.TestResourceDataRaw(t, Schema(), map[string]interface{}{
				constants.FieldCommonName:               name,
				constants.FieldKubeOVNVpcEnableExternal: tc.value,
				constants.FieldKubeOVNVpcEnableBfd:      tc.value,
			})
			constructor := Creator(name)
			if tc.existing != nil {
				constructor = Updater(tc.existing)
			}
			obj, err := util.ResourceConstruct(ctx, d, constructor)
			if err != nil {
				t.Fatal(err)
			}
			spec := obj.(*kubeovnv1.Vpc).Spec
			if spec.EnableExternal != tc.value || spec.EnableBfd != tc.value {
				t.Errorf("enable_external = %v, enable_bfd = %v, want %v", spec.EnableExternal, spec.EnableBfd, tc.value)
			}
		})
	}
}
