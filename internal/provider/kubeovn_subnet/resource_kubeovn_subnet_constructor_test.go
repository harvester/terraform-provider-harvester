package kubeovn_subnet

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

func TestSubnetConstructor(t *testing.T) {
	const (
		namespace1 = "ns1"
		namespace2 = "ns2"
		cidr1      = "10.1.0.0/16"
		cidr2      = "10.2.0.0/16"
	)

	existing := &kubeovnv1.Subnet{
		ObjectMeta: metav1.ObjectMeta{Name: subnetName},
		Spec: kubeovnv1.SubnetSpec{
			CIDRBlock:    subnetCIDR,
			Gateway:      gatewayIP,
			ExcludeIps:   []string{gatewayIP, "10.0.0.99"},
			Namespaces:   []string{"old-ns"},
			AllowSubnets: []string{"192.168.0.0/16"},
		},
	}

	testcases := []struct {
		name                 string
		existing             *kubeovnv1.Subnet // nil creates the subnet
		config               map[string]interface{}
		expectedExcludeIPs   []string
		expectedNamespaces   []string
		expectedAllowSubnets []string
	}{
		{
			name:   "create without lists",
			config: map[string]interface{}{},
		},
		{
			name: "create sorts exclude_ips and keeps the order of the other lists",
			config: map[string]interface{}{
				constants.FieldKubeOVNSubnetExcludeIPs:   []interface{}{otherExcludedIP, gatewayIP, excludedRange},
				constants.FieldKubeOVNSubnetNamespaces:   []interface{}{namespace2, namespace1, "ns3"},
				constants.FieldKubeOVNSubnetAllowSubnets: []interface{}{cidr2, cidr1},
			},
			expectedExcludeIPs:   []string{gatewayIP, excludedRange, otherExcludedIP},
			expectedNamespaces:   []string{namespace2, namespace1, "ns3"},
			expectedAllowSubnets: []string{cidr2, cidr1},
		},
		{
			name: "empty exclude_ips entries are dropped",
			config: map[string]interface{}{
				constants.FieldKubeOVNSubnetExcludeIPs: []interface{}{"", excludedIP},
			},
			expectedExcludeIPs: []string{excludedIP},
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
				constants.FieldKubeOVNSubnetExcludeIPs:   []interface{}{excludedIP, gatewayIP},
				constants.FieldKubeOVNSubnetNamespaces:   []interface{}{namespace1, namespace2},
				constants.FieldKubeOVNSubnetAllowSubnets: []interface{}{cidr1, cidr2},
			},
			expectedExcludeIPs:   []string{gatewayIP, excludedIP},
			expectedNamespaces:   []string{namespace1, namespace2},
			expectedAllowSubnets: []string{cidr1, cidr2},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			constructor := Creator(subnetName)
			if tc.existing != nil {
				constructor = Updater(tc.existing.DeepCopy())
			}
			d := schema.TestResourceDataRaw(t, Schema(), tc.config)
			result, err := util.ResourceConstruct(context.Background(), d, constructor)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			subnet := result.(*kubeovnv1.Subnet)

			if !slices.Equal(subnet.Spec.ExcludeIps, tc.expectedExcludeIPs) {
				t.Errorf("ExcludeIps: expected %v, got %v", tc.expectedExcludeIPs, subnet.Spec.ExcludeIps)
			}
			if !slices.Equal(subnet.Spec.Namespaces, tc.expectedNamespaces) {
				t.Errorf("Namespaces: expected %v, got %v", tc.expectedNamespaces, subnet.Spec.Namespaces)
			}
			if !slices.Equal(subnet.Spec.AllowSubnets, tc.expectedAllowSubnets) {
				t.Errorf("AllowSubnets: expected %v, got %v", tc.expectedAllowSubnets, subnet.Spec.AllowSubnets)
			}
		})
	}
}

// TestSubnetConstructorBooleans checks that every boolean is sent as
// configured, false included, so that an update can switch it off.
func TestSubnetConstructorBooleans(t *testing.T) {
	existing := func() *kubeovnv1.Subnet {
		enableLb := true
		return &kubeovnv1.Subnet{
			ObjectMeta: metav1.ObjectMeta{Name: subnetName},
			Spec: kubeovnv1.SubnetSpec{
				CIDRBlock:   subnetCIDR,
				Gateway:     gatewayIP,
				EnableDHCP:  true,
				Private:     true,
				NatOutgoing: true,
				EnableLb:    &enableLb,
			},
		}
	}
	testcases := []struct {
		name     string
		existing *kubeovnv1.Subnet // nil creates the subnet
		value    bool
	}{
		{name: "create with true", value: true},
		{name: "create with false", value: false},
		{name: "update switches every boolean off", existing: existing(), value: false},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			d := schema.TestResourceDataRaw(t, Schema(), map[string]interface{}{
				constants.FieldCommonName:               subnetName,
				constants.FieldKubeOVNSubnetCIDRBlock:   subnetCIDR,
				constants.FieldKubeOVNSubnetGateway:     gatewayIP,
				constants.FieldKubeOVNSubnetEnableDHCP:  tc.value,
				constants.FieldKubeOVNSubnetPrivate:     tc.value,
				constants.FieldKubeOVNSubnetNatOutgoing: tc.value,
				constants.FieldKubeOVNSubnetEnableLb:    tc.value,
			})
			constructor := Creator(subnetName)
			if tc.existing != nil {
				constructor = Updater(tc.existing)
			}
			obj, err := util.ResourceConstruct(ctx, d, constructor)
			if err != nil {
				t.Fatal(err)
			}
			spec := obj.(*kubeovnv1.Subnet).Spec
			got := map[string]bool{"enable_dhcp": spec.EnableDHCP, "private": spec.Private, "nat_outgoing": spec.NatOutgoing}
			for field, value := range got {
				if value != tc.value {
					t.Errorf("%s = %v, want %v", field, value, tc.value)
				}
			}
			// A missing enable_lb takes the --enable-lb setting of the kube-ovn
			// controller, so it must always be sent.
			if spec.EnableLb == nil || *spec.EnableLb != tc.value {
				t.Errorf("enable_lb = %v, want %v", spec.EnableLb, tc.value)
			}
		})
	}
}

// TestSubnetConstructorOptionalStrings checks that removing vlan,
// network_provider, dhcp_v4_options or gateway_node from the configuration
// clears them: the updater starts from the live object, so a field that is
// not sent keeps its old value.
func TestSubnetConstructorOptionalStrings(t *testing.T) {
	const (
		vlan        = "vlan100"
		provider    = "net1.default.ovn"
		dhcpOptions = "lease_time=3600"
		gatewayNode = "node1,node2"
	)
	existing := func() *kubeovnv1.Subnet {
		return &kubeovnv1.Subnet{
			ObjectMeta: metav1.ObjectMeta{Name: subnetName},
			Spec: kubeovnv1.SubnetSpec{
				CIDRBlock:     subnetCIDR,
				Gateway:       gatewayIP,
				Vlan:          "old-vlan",
				Provider:      "old.default.ovn",
				DHCPv4Options: "lease_time=60",
				GatewayNode:   "old-node",
			},
		}
	}
	configured := map[string]string{
		constants.FieldKubeOVNSubnetVlan:          vlan,
		constants.FieldKubeOVNSubnetProvider:      provider,
		constants.FieldKubeOVNSubnetDHCPv4Options: dhcpOptions,
		constants.FieldKubeOVNSubnetGatewayNode:   gatewayNode,
	}
	empty := map[string]string{}

	testcases := []struct {
		name     string
		existing *kubeovnv1.Subnet // nil creates the subnet
		config   map[string]string
	}{
		{name: "create without the fields", config: empty},
		{name: "create with every field", config: configured},
		{name: "update replaces every field", existing: existing(), config: configured},
		{name: "update clears the fields removed from the configuration", existing: existing(), config: empty},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			raw := map[string]interface{}{
				constants.FieldCommonName:             subnetName,
				constants.FieldKubeOVNSubnetCIDRBlock: subnetCIDR,
				constants.FieldKubeOVNSubnetGateway:   gatewayIP,
			}
			for field, value := range tc.config {
				raw[field] = value
			}
			constructor := Creator(subnetName)
			if tc.existing != nil {
				constructor = Updater(tc.existing)
			}
			obj, err := util.ResourceConstruct(context.Background(), schema.TestResourceDataRaw(t, Schema(), raw), constructor)
			if err != nil {
				t.Fatal(err)
			}
			spec := obj.(*kubeovnv1.Subnet).Spec
			got := map[string]string{
				constants.FieldKubeOVNSubnetVlan:          spec.Vlan,
				constants.FieldKubeOVNSubnetProvider:      spec.Provider,
				constants.FieldKubeOVNSubnetDHCPv4Options: spec.DHCPv4Options,
				constants.FieldKubeOVNSubnetGatewayNode:   spec.GatewayNode,
			}
			for field, value := range got {
				if value != tc.config[field] {
					t.Errorf("%s = %q, want %q", field, value, tc.config[field])
				}
			}
		})
	}
}
