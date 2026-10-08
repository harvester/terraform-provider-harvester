package importer

import (
	"slices"
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceKubeOVNSecurityGroupStateGetter(t *testing.T) {
	const allAddresses = "0.0.0.0/0"

	testcases := []struct {
		name             string
		sg               *kubeovnv1.SecurityGroup
		expectedID       string
		expectedAllow    bool
		expectedIngress  []string // remote addresses, in order
		expectedEgress   []string // remote addresses, in order
		expectedPolicies []string // ingress policies, in order
	}{
		{
			name: "security group with rules",
			sg: &kubeovnv1.SecurityGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-sg",
				},
				Spec: kubeovnv1.SecurityGroupSpec{
					AllowSameGroupTraffic: true,
					IngressRules: []kubeovnv1.SecurityGroupRule{
						{
							IPVersion:     "ipv4",
							Protocol:      kubeovnv1.SgProtocolTCP,
							Priority:      1,
							RemoteType:    kubeovnv1.SgRemoteTypeAddress,
							RemoteAddress: testCIDR,
							PortRangeMin:  80,
							PortRangeMax:  80,
							Policy:        kubeovnv1.SgPolicyAllow,
						},
					},
					EgressRules: []kubeovnv1.SecurityGroupRule{
						{
							IPVersion:     "ipv4",
							Protocol:      kubeovnv1.SgProtocolALL,
							RemoteType:    kubeovnv1.SgRemoteTypeAddress,
							RemoteAddress: allAddresses,
							Policy:        kubeovnv1.SgPolicyAllow,
						},
					},
				},
				Status: kubeovnv1.SecurityGroupStatus{
					PortGroup:              "test-sg-pg",
					IngressLastSyncSuccess: true,
					EgressLastSyncSuccess:  true,
				},
			},
			expectedID:       helper.BuildID("", "test-sg"),
			expectedAllow:    true,
			expectedIngress:  []string{testCIDR},
			expectedEgress:   []string{allAddresses},
			expectedPolicies: []string{string(kubeovnv1.SgPolicyAllow)},
		},
		{
			name: "several rules keep their order",
			sg: &kubeovnv1.SecurityGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ordered-sg",
				},
				Spec: kubeovnv1.SecurityGroupSpec{
					IngressRules: []kubeovnv1.SecurityGroupRule{
						{Priority: 3, RemoteAddress: "10.3.0.0/16", Policy: kubeovnv1.SgPolicyDrop},
						{Priority: 1, RemoteAddress: testCIDR1, Policy: kubeovnv1.SgPolicyAllow},
						{Priority: 2, RemoteAddress: testCIDR2, Policy: kubeovnv1.SgPolicyDrop},
					},
					EgressRules: []kubeovnv1.SecurityGroupRule{
						{RemoteAddress: allAddresses, Policy: kubeovnv1.SgPolicyAllow},
						{RemoteAddress: testCIDR, Policy: kubeovnv1.SgPolicyDrop},
					},
				},
			},
			expectedID:       helper.BuildID("", "ordered-sg"),
			expectedIngress:  []string{"10.3.0.0/16", testCIDR1, testCIDR2},
			expectedEgress:   []string{allAddresses, testCIDR},
			expectedPolicies: []string{string(kubeovnv1.SgPolicyDrop), string(kubeovnv1.SgPolicyAllow), string(kubeovnv1.SgPolicyDrop)},
		},
		{
			name: "empty security group",
			sg: &kubeovnv1.SecurityGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name: "empty-sg",
				},
				Spec: kubeovnv1.SecurityGroupSpec{},
			},
			expectedID:    helper.BuildID("", "empty-sg"),
			expectedAllow: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNSecurityGroupStateGetter(tc.sg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNSecurityGroup {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeKubeOVNSecurityGroup, getter.ResourceType)
			}

			allow := getter.States[constants.FieldKubeOVNSGAllowSameGroupTraffic].(bool)
			if allow != tc.expectedAllow {
				t.Errorf("AllowSameGroupTraffic: expected %v, got %v", tc.expectedAllow, allow)
			}

			ingress := blockStrings(getter.States[constants.FieldKubeOVNSGIngressRules], constants.FieldKubeOVNSGRuleRemoteAddress)
			if !slices.Equal(ingress, tc.expectedIngress) {
				t.Errorf("IngressRules: expected %v, got %v", tc.expectedIngress, ingress)
			}

			egress := blockStrings(getter.States[constants.FieldKubeOVNSGEgressRules], constants.FieldKubeOVNSGRuleRemoteAddress)
			if !slices.Equal(egress, tc.expectedEgress) {
				t.Errorf("EgressRules: expected %v, got %v", tc.expectedEgress, egress)
			}

			policies := blockStrings(getter.States[constants.FieldKubeOVNSGIngressRules], constants.FieldKubeOVNSGRulePolicy)
			if !slices.Equal(policies, tc.expectedPolicies) {
				t.Errorf("IngressRule policies: expected %v, got %v", tc.expectedPolicies, policies)
			}
		})
	}
}
