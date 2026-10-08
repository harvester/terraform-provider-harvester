package kubeovn_security_group

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

const (
	sgName = "test-sg"
)

func sgRule(remoteAddress, policy string) map[string]interface{} {
	return map[string]interface{}{
		constants.FieldKubeOVNSGRuleIPVersion:     "ipv4",
		constants.FieldKubeOVNSGRulePriority:      1,
		constants.FieldKubeOVNSGRuleRemoteType:    string(kubeovnv1.SgRemoteTypeAddress),
		constants.FieldKubeOVNSGRuleRemoteAddress: remoteAddress,
		constants.FieldKubeOVNSGRulePolicy:        policy,
	}
}

// remoteAddresses returns the remote address of every rule, in order.
func remoteAddresses(rules []kubeovnv1.SecurityGroupRule) []string {
	addresses := make([]string, 0, len(rules))
	for _, rule := range rules {
		addresses = append(addresses, rule.RemoteAddress)
	}
	return addresses
}

func TestSecurityGroupConstructor(t *testing.T) {
	const (
		cidr1 = "10.1.0.0/16"
		cidr2 = "10.2.0.0/16"
	)
	allow, drop := string(kubeovnv1.SgPolicyAllow), string(kubeovnv1.SgPolicyDrop)

	existing := &kubeovnv1.SecurityGroup{
		ObjectMeta: metav1.ObjectMeta{Name: sgName},
		Spec: kubeovnv1.SecurityGroupSpec{
			IngressRules: []kubeovnv1.SecurityGroupRule{{RemoteAddress: "192.168.0.0/16"}},
			EgressRules:  []kubeovnv1.SecurityGroupRule{{RemoteAddress: "192.168.0.0/16"}},
		},
	}

	testcases := []struct {
		name            string
		existing        *kubeovnv1.SecurityGroup // nil creates the security group
		config          map[string]interface{}
		expectedIngress []string // remote addresses, in order
		expectedEgress  []string // remote addresses, in order
	}{
		{
			name: "create keeps the order of several rules",
			config: map[string]interface{}{
				constants.FieldKubeOVNSGIngressRules: []interface{}{sgRule(cidr2, drop), sgRule(cidr1, allow), sgRule("0.0.0.0/0", drop)},
				constants.FieldKubeOVNSGEgressRules:  []interface{}{sgRule(cidr1, allow), sgRule(cidr2, drop)},
			},
			expectedIngress: []string{cidr2, cidr1, "0.0.0.0/0"},
			expectedEgress:  []string{cidr1, cidr2},
		},
		{
			name:     "update drops rules removed from the configuration",
			existing: existing,
			config:   map[string]interface{}{},
		},
		{
			name:     "update replaces rules instead of appending to them",
			existing: existing,
			config: map[string]interface{}{
				constants.FieldKubeOVNSGIngressRules: []interface{}{sgRule(cidr1, allow), sgRule(cidr2, allow)},
				constants.FieldKubeOVNSGEgressRules:  []interface{}{sgRule(cidr2, drop)},
			},
			expectedIngress: []string{cidr1, cidr2},
			expectedEgress:  []string{cidr2},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			constructor := Creator(sgName)
			if tc.existing != nil {
				constructor = Updater(tc.existing.DeepCopy())
			}
			d := schema.TestResourceDataRaw(t, Schema(), tc.config)
			result, err := util.ResourceConstruct(context.Background(), d, constructor)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			sg := result.(*kubeovnv1.SecurityGroup)

			if ingress := remoteAddresses(sg.Spec.IngressRules); !slices.Equal(ingress, tc.expectedIngress) {
				t.Errorf("IngressRules: expected %v, got %v", tc.expectedIngress, ingress)
			}
			if egress := remoteAddresses(sg.Spec.EgressRules); !slices.Equal(egress, tc.expectedEgress) {
				t.Errorf("EgressRules: expected %v, got %v", tc.expectedEgress, egress)
			}
		})
	}
}

// TestSecurityGroupConstructorAllowSameGroupTraffic checks that allow_same_group_traffic is sent as configured, false included, so
// that an update can switch it off.
func TestSecurityGroupConstructorAllowSameGroupTraffic(t *testing.T) {
	testcases := []struct {
		name     string
		existing *kubeovnv1.SecurityGroup // nil creates the object
		value    bool
	}{
		{name: "create with true", value: true},
		{name: "create with false", value: false},
		{name: "update switches it off", existing: &kubeovnv1.SecurityGroup{ObjectMeta: metav1.ObjectMeta{Name: sgName}, Spec: kubeovnv1.SecurityGroupSpec{AllowSameGroupTraffic: true}}, value: false},
		{name: "update switches it on", existing: &kubeovnv1.SecurityGroup{ObjectMeta: metav1.ObjectMeta{Name: sgName}}, value: true},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, Schema(), map[string]interface{}{
				constants.FieldCommonName:                     sgName,
				constants.FieldKubeOVNSGAllowSameGroupTraffic: tc.value,
			})
			constructor := Creator(sgName)
			if tc.existing != nil {
				constructor = Updater(tc.existing)
			}
			obj, err := util.ResourceConstruct(context.Background(), d, constructor)
			if err != nil {
				t.Fatal(err)
			}
			if got := obj.(*kubeovnv1.SecurityGroup).Spec.AllowSameGroupTraffic; got != tc.value {
				t.Errorf("allow_same_group_traffic = %v, want %v", got, tc.value)
			}
		})
	}
}
