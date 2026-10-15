package kubeovn_security_group

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// kube-ovn skips the whole security group when a rule priority is out of
// range, and an omitted priority is sent as 0: the plan must refuse both.
func TestSecurityGroupRulePriority(t *testing.T) {
	const omitted = -1

	testcases := []struct {
		name        string
		priorities  []int // one ingress rule per entry, omitted leaves it out
		expectedErr bool
	}{
		{name: "lowest priority", priorities: []int{1}},
		{name: "several valid rules", priorities: []int{1, 100, 200}},
		{name: "priority above 200 is left to the kube-ovn version", priorities: []int{16384}},
		{name: "omitted priority", priorities: []int{omitted}, expectedErr: true},
		{name: "zero priority", priorities: []int{0}, expectedErr: true},
		{name: "one invalid rule among valid ones", priorities: []int{1, -5, 2}, expectedErr: true},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			rules := make([]interface{}, 0, len(tc.priorities))
			for _, priority := range tc.priorities {
				rule := sgRule("10.0.0.0/24", "allow")
				if priority == omitted {
					delete(rule, constants.FieldKubeOVNSGRulePriority)
				} else {
					rule[constants.FieldKubeOVNSGRulePriority] = priority
				}
				rules = append(rules, rule)
			}
			diags := ResourceKubeOVNSecurityGroup().Validate(terraform.NewResourceConfigRaw(map[string]interface{}{
				constants.FieldCommonName:            sgName,
				constants.FieldKubeOVNSGIngressRules: rules,
			}))
			if tc.expectedErr != diags.HasError() {
				t.Errorf("diagnostics = %v, want an error: %v", diags, tc.expectedErr)
			}
		})
	}
}
