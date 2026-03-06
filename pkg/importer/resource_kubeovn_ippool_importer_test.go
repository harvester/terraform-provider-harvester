package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceKubeOVNIPPoolStateGetter(t *testing.T) {
	testcases := []struct {
		name           string
		ippool         *kubeovnv1.IPPool
		expectedID     string
		expectedSubnet string
		expectedIPs    int
		expectedNS     int
	}{
		{
			name: "ippool with ips and namespaces",
			ippool: &kubeovnv1.IPPool{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool",
					Labels: map[string]string{
						"tag.harvesterhci.io/env": "test",
					},
					Annotations: map[string]string{
						"field.cattle.io/description": "Test pool",
					},
				},
				Spec: kubeovnv1.IPPoolSpec{
					Subnet:     "test-subnet",
					IPs:        []string{"10.0.0.10", "10.0.0.20..10.0.0.30"},
					Namespaces: []string{"ns1", "ns2"},
				},
			},
			expectedID:     helper.BuildID("", "test-pool"),
			expectedSubnet: "test-subnet",
			expectedIPs:    2,
			expectedNS:     2,
		},
		{
			name: "empty ippool",
			ippool: &kubeovnv1.IPPool{
				ObjectMeta: metav1.ObjectMeta{
					Name: "empty-pool",
				},
				Spec: kubeovnv1.IPPoolSpec{},
			},
			expectedID:     helper.BuildID("", "empty-pool"),
			expectedSubnet: "",
			expectedIPs:    0,
			expectedNS:     0,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNIPPoolStateGetter(tc.ippool)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.Name != tc.ippool.Name {
				t.Errorf("Name: expected %q, got %q", tc.ippool.Name, getter.Name)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNIPPool {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeKubeOVNIPPool, getter.ResourceType)
			}

			subnet := getter.States[constants.FieldKubeOVNIPPoolSubnet].(string)
			if subnet != tc.expectedSubnet {
				t.Errorf("Subnet: expected %q, got %q", tc.expectedSubnet, subnet)
			}

			ips := getter.States[constants.FieldKubeOVNIPPoolIPs].([]string)
			if len(ips) != tc.expectedIPs {
				t.Errorf("IPs: expected %d, got %d", tc.expectedIPs, len(ips))
			}

			var nsLen int
			if ns := getter.States[constants.FieldKubeOVNIPPoolNamespaces]; ns != nil {
				nsLen = len(ns.([]string))
			}
			if nsLen != tc.expectedNS {
				t.Errorf("Namespaces: expected %d, got %d", tc.expectedNS, nsLen)
			}
		})
	}
}
