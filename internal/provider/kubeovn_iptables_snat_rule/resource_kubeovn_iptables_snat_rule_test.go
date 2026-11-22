package kubeovn_iptables_snat_rule

import (
	"context"
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	kubeovnfake "github.com/kubeovn/kube-ovn/pkg/client/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func TestSnatRuleRefresh(t *testing.T) {
	const name = "test-snat"

	testcases := []struct {
		name          string
		obj           *kubeovnv1.IptablesSnatRule // nil when the object is gone
		expectedState string
	}{
		{
			name:          "deleted",
			expectedState: constants.StateCommonRemoved,
		},
		{
			name:          "not applied yet",
			obj:           &kubeovnv1.IptablesSnatRule{ObjectMeta: metav1.ObjectMeta{Name: name}},
			expectedState: constants.StateCommonActive,
		},
		{
			name:          "applied in the NAT gateway",
			obj:           &kubeovnv1.IptablesSnatRule{ObjectMeta: metav1.ObjectMeta{Name: name}, Status: kubeovnv1.IptablesSnatRuleStatus{Ready: true}},
			expectedState: constants.StateCommonReady,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// The object is created through the typed client: the fake tracker
			// cannot derive the hyphenated kube-ovn resource name from the kind.
			objects := kubeovnfake.NewSimpleClientset().KubeovnV1().IptablesSnatRules()
			if tc.obj != nil {
				if _, err := objects.Create(context.Background(), tc.obj, metav1.CreateOptions{}); err != nil {
					t.Fatal(err)
				}
			}
			_, state, err := snatRuleRefresh(context.Background(), objects, name)()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if state != tc.expectedState {
				t.Errorf("state = %q, want %q", state, tc.expectedState)
			}
		})
	}
}
