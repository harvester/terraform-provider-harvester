package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceKubeOVNSwitchLBRuleStateGetter(t *testing.T) {
	testcases := []struct {
		name          string
		slr           *kubeovnv1.SwitchLBRule
		expectedID    string
		expectedVip   string
		expectedPorts int
		expectedState string
	}{
		{
			name: "switch lb rule with ports",
			slr: &kubeovnv1.SwitchLBRule{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-slr",
				},
				Spec: kubeovnv1.SwitchLBRuleSpec{
					Vip:       "10.10.0.1",
					Namespace: "default",
					Selector:  []string{"app=web"},
					Endpoints: []string{"10.0.0.1", "10.0.0.2"},
					Ports: []kubeovnv1.SlrPort{
						{
							Name:       "http",
							Port:       80,
							TargetPort: 8080,
							Protocol:   "TCP",
						},
					},
				},
				Status: kubeovnv1.SwitchLBRuleStatus{
					Ports:   "http:80/TCP",
					Service: "test-slr-svc",
				},
			},
			expectedID:    helper.BuildID("", "test-slr"),
			expectedVip:   "10.10.0.1",
			expectedPorts: 1,
			expectedState: constants.StateCommonReady,
		},
		{
			name: "empty switch lb rule",
			slr: &kubeovnv1.SwitchLBRule{
				ObjectMeta: metav1.ObjectMeta{
					Name: "empty-slr",
				},
				Spec: kubeovnv1.SwitchLBRuleSpec{},
			},
			expectedID:    helper.BuildID("", "empty-slr"),
			expectedVip:   "",
			expectedPorts: 0,
			expectedState: constants.StateCommonActive,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNSwitchLBRuleStateGetter(tc.slr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNSwitchLBRule {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeKubeOVNSwitchLBRule, getter.ResourceType)
			}

			vip := getter.States[constants.FieldKubeOVNSwitchLBRuleVip].(string)
			if vip != tc.expectedVip {
				t.Errorf("Vip: expected %q, got %q", tc.expectedVip, vip)
			}

			ports := getter.States[constants.FieldKubeOVNSwitchLBRulePorts].([]map[string]interface{})
			if len(ports) != tc.expectedPorts {
				t.Errorf("Ports: expected %d, got %d", tc.expectedPorts, len(ports))
			}

			state := getter.States[constants.FieldCommonState].(string)
			if state != tc.expectedState {
				t.Errorf("State: expected %q, got %q", tc.expectedState, state)
			}

			if tc.expectedPorts > 0 {
				port := ports[0]
				if port[constants.FieldKubeOVNSlrPortName] != "http" {
					t.Errorf("Port Name: expected %q, got %q", "http", port[constants.FieldKubeOVNSlrPortName])
				}
				if port[constants.FieldKubeOVNSlrPortPort] != 80 {
					t.Errorf("Port Port: expected %d, got %v", 80, port[constants.FieldKubeOVNSlrPortPort])
				}
				if port[constants.FieldKubeOVNSlrPortProtocol] != "TCP" {
					t.Errorf("Port Protocol: expected %q, got %q", "TCP", port[constants.FieldKubeOVNSlrPortProtocol])
				}
			}
		})
	}
}
