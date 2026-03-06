package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceKubeOVNVipStateGetter(t *testing.T) {
	testcases := []struct {
		name                  string
		vip                   *kubeovnv1.Vip
		expectedID            string
		expectedState         string
		expectedSubnet        string
		expectedV4IP          string
		expectedType          string
		expectedReady         bool
		expectedStatusV4IP    string
		expectedSelector      []string
		expectedAttachSubnets []string
	}{
		{
			name: "vip with all fields",
			vip: &kubeovnv1.Vip{
				ObjectMeta: metav1.ObjectMeta{Name: "test-vip"},
				Spec: kubeovnv1.VipSpec{
					Namespace:     "default",
					Subnet:        "test-subnet",
					Type:          "allowed_address_pair",
					V4ip:          "10.0.0.100",
					V6ip:          "fd00::100",
					MacAddress:    "00:11:22:33:44:55",
					ParentV4ip:    "10.0.0.1",
					ParentV6ip:    "fd00::1",
					ParentMac:     "00:11:22:33:44:00",
					Selector:      []string{"a", "b"},
					AttachSubnets: []string{"sub1"},
				},
				Status: kubeovnv1.VipStatus{
					Ready: true,
					V4ip:  "10.0.0.100",
					V6ip:  "fd00::100",
					Mac:   "00:11:22:33:44:55",
					Type:  "allowed_address_pair",
				},
			},
			expectedID:            helper.BuildID("", "test-vip"),
			expectedState:         constants.StateCommonReady,
			expectedSubnet:        "test-subnet",
			expectedV4IP:          "10.0.0.100",
			expectedType:          "allowed_address_pair",
			expectedReady:         true,
			expectedStatusV4IP:    "10.0.0.100",
			expectedSelector:      []string{"a", "b"},
			expectedAttachSubnets: []string{"sub1"},
		},
		{
			name: "empty vip",
			vip: &kubeovnv1.Vip{
				ObjectMeta: metav1.ObjectMeta{Name: "empty-vip"},
			},
			expectedID:            helper.BuildID("", "empty-vip"),
			expectedState:         constants.StateCommonActive,
			expectedSubnet:        "",
			expectedV4IP:          "",
			expectedType:          "",
			expectedReady:         false,
			expectedStatusV4IP:    "",
			expectedSelector:      nil,
			expectedAttachSubnets: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNVipStateGetter(tc.vip)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNVip {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeKubeOVNVip, getter.ResourceType)
			}
			state := getter.States[constants.FieldCommonState].(string)
			if state != tc.expectedState {
				t.Errorf("State: expected %q, got %q", tc.expectedState, state)
			}
			subnet := getter.States[constants.FieldKubeOVNVipSubnet].(string)
			if subnet != tc.expectedSubnet {
				t.Errorf("Subnet: expected %q, got %q", tc.expectedSubnet, subnet)
			}
			v4ip := getter.States[constants.FieldKubeOVNVipV4IP].(string)
			if v4ip != tc.expectedV4IP {
				t.Errorf("V4IP: expected %q, got %q", tc.expectedV4IP, v4ip)
			}
			vipType := getter.States[constants.FieldKubeOVNVipType].(string)
			if vipType != tc.expectedType {
				t.Errorf("Type: expected %q, got %q", tc.expectedType, vipType)
			}
			ready := getter.States[constants.FieldKubeOVNVipStatusReady].(bool)
			if ready != tc.expectedReady {
				t.Errorf("Ready: expected %v, got %v", tc.expectedReady, ready)
			}
			statusV4IP := getter.States[constants.FieldKubeOVNVipStatusV4IP].(string)
			if statusV4IP != tc.expectedStatusV4IP {
				t.Errorf("StatusV4IP: expected %q, got %q", tc.expectedStatusV4IP, statusV4IP)
			}
			selector := getter.States[constants.FieldKubeOVNVipSelector]
			if tc.expectedSelector == nil {
				if selector != nil {
					selectorSlice, ok := selector.([]string)
					if ok && len(selectorSlice) > 0 {
						t.Errorf("Selector: expected nil, got %v", selector)
					}
				}
			} else {
				selectorSlice := selector.([]string)
				if len(selectorSlice) != len(tc.expectedSelector) {
					t.Errorf("Selector length: expected %d, got %d", len(tc.expectedSelector), len(selectorSlice))
				} else {
					for i, v := range tc.expectedSelector {
						if selectorSlice[i] != v {
							t.Errorf("Selector[%d]: expected %q, got %q", i, v, selectorSlice[i])
						}
					}
				}
			}
			attachSubnets := getter.States[constants.FieldKubeOVNVipAttachSubnets]
			if tc.expectedAttachSubnets == nil {
				if attachSubnets != nil {
					attachSlice, ok := attachSubnets.([]string)
					if ok && len(attachSlice) > 0 {
						t.Errorf("AttachSubnets: expected nil, got %v", attachSubnets)
					}
				}
			} else {
				attachSlice := attachSubnets.([]string)
				if len(attachSlice) != len(tc.expectedAttachSubnets) {
					t.Errorf("AttachSubnets length: expected %d, got %d", len(tc.expectedAttachSubnets), len(attachSlice))
				} else {
					for i, v := range tc.expectedAttachSubnets {
						if attachSlice[i] != v {
							t.Errorf("AttachSubnets[%d]: expected %q, got %q", i, v, attachSlice[i])
						}
					}
				}
			}
		})
	}
}
