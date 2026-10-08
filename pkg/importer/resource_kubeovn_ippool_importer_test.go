package importer

import (
	"slices"
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
		expectedIPs    []string
		expectedNS     []string
	}{
		{
			name: "ippool with ips and namespaces",
			ippool: &kubeovnv1.IPPool{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pool",
					Labels: map[string]string{
						testTagKey: testTagValue,
					},
					Annotations: map[string]string{
						testDescriptionKey: "Test pool",
					},
				},
				Spec: kubeovnv1.IPPoolSpec{
					Subnet:     "test-subnet",
					IPs:        []string{"10.0.0.10", "10.0.0.20..10.0.0.30"},
					Namespaces: []string{testNamespace1, testNamespace2},
				},
			},
			expectedID:     helper.BuildID("", "test-pool"),
			expectedSubnet: "test-subnet",
			expectedIPs:    []string{"10.0.0.10", "10.0.0.20..10.0.0.30"},
			expectedNS:     []string{testNamespace1, testNamespace2},
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
			if !slices.Equal(ips, tc.expectedIPs) {
				t.Errorf("IPs: expected %v, got %v", tc.expectedIPs, ips)
			}

			namespaces := getter.States[constants.FieldKubeOVNIPPoolNamespaces].([]string)
			if !slices.Equal(namespaces, tc.expectedNS) {
				t.Errorf("Namespaces: expected %v, got %v", tc.expectedNS, namespaces)
			}
		})
	}
}
