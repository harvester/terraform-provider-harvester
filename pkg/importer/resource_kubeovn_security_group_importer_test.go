package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceKubeOVNSecurityGroupStateGetter(t *testing.T) {
	testcases := []struct {
		name            string
		sg              *kubeovnv1.SecurityGroup
		expectedID      string
		expectedAllow   bool
		expectedIngress int
		expectedEgress  int
	}{
		{
			name: "security group with rules",
			sg: &kubeovnv1.SecurityGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-sg",
				},
				Spec: kubeovnv1.SecurityGroupSpec{
					AllowSameGroupTraffic: true,
					IngressRules: []*kubeovnv1.SgRule{
						{
							IPVersion:     "ipv4",
							Protocol:      "tcp",
							Priority:      1,
							RemoteType:    "address",
							RemoteAddress: "10.0.0.0/24",
							PortRangeMin:  80,
							PortRangeMax:  80,
							Policy:        "allow",
						},
					},
					EgressRules: []*kubeovnv1.SgRule{
						{
							IPVersion:     "ipv4",
							Protocol:      "all",
							RemoteType:    "address",
							RemoteAddress: "0.0.0.0/0",
							Policy:        "allow",
						},
					},
				},
				Status: kubeovnv1.SecurityGroupStatus{
					PortGroup:              "test-sg-pg",
					IngressLastSyncSuccess: true,
					EgressLastSyncSuccess:  true,
				},
			},
			expectedID:      helper.BuildID("", "test-sg"),
			expectedAllow:   true,
			expectedIngress: 1,
			expectedEgress:  1,
		},
		{
			name: "empty security group",
			sg: &kubeovnv1.SecurityGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name: "empty-sg",
				},
				Spec: kubeovnv1.SecurityGroupSpec{},
			},
			expectedID:      helper.BuildID("", "empty-sg"),
			expectedAllow:   false,
			expectedIngress: 0,
			expectedEgress:  0,
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

			ingress := getter.States[constants.FieldKubeOVNSGIngressRules].([]map[string]interface{})
			if len(ingress) != tc.expectedIngress {
				t.Errorf("IngressRules: expected %d, got %d", tc.expectedIngress, len(ingress))
			}

			egress := getter.States[constants.FieldKubeOVNSGEgressRules].([]map[string]interface{})
			if len(egress) != tc.expectedEgress {
				t.Errorf("EgressRules: expected %d, got %d", tc.expectedEgress, len(egress))
			}

			if tc.expectedIngress > 0 {
				rule := ingress[0]
				if rule[constants.FieldKubeOVNSGRuleIPVersion] != "ipv4" {
					t.Errorf("IngressRule IPVersion: expected %q, got %q", "ipv4", rule[constants.FieldKubeOVNSGRuleIPVersion])
				}
				if rule[constants.FieldKubeOVNSGRulePolicy] != "allow" {
					t.Errorf("IngressRule Policy: expected %q, got %q", "allow", rule[constants.FieldKubeOVNSGRulePolicy])
				}
			}
		})
	}
}
