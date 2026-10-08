package kubeovn_ippool

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

func TestIPPoolConstructor(t *testing.T) {
	const (
		singleIP   = "10.0.0.10"
		ipRange    = "10.0.0.20..10.0.0.30"
		namespace1 = "ns1"
		namespace2 = "ns2"
	)

	existing := &kubeovnv1.IPPool{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pool"},
		Spec: kubeovnv1.IPPoolSpec{
			IPs:        []string{"10.0.0.99"},
			Namespaces: []string{"old-ns"},
		},
	}

	testcases := []struct {
		name               string
		existing           *kubeovnv1.IPPool // nil creates the pool
		config             map[string]interface{}
		expectedIPs        []string
		expectedNamespaces []string
	}{
		{
			name: "create keeps the order of several entries",
			config: map[string]interface{}{
				constants.FieldKubeOVNIPPoolIPs:        []interface{}{ipRange, singleIP, "10.0.0.0/28"},
				constants.FieldKubeOVNIPPoolNamespaces: []interface{}{namespace2, namespace1},
			},
			expectedIPs:        []string{ipRange, singleIP, "10.0.0.0/28"},
			expectedNamespaces: []string{namespace2, namespace1},
		},
		{
			name:     "update drops namespaces removed from the configuration",
			existing: existing,
			config: map[string]interface{}{
				constants.FieldKubeOVNIPPoolIPs: []interface{}{singleIP},
			},
			expectedIPs: []string{singleIP},
		},
		{
			name:     "update replaces lists instead of appending to them",
			existing: existing,
			config: map[string]interface{}{
				constants.FieldKubeOVNIPPoolIPs:        []interface{}{singleIP, ipRange},
				constants.FieldKubeOVNIPPoolNamespaces: []interface{}{namespace1, namespace2},
			},
			expectedIPs:        []string{singleIP, ipRange},
			expectedNamespaces: []string{namespace1, namespace2},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			constructor := Creator("test-pool")
			if tc.existing != nil {
				constructor = Updater(tc.existing.DeepCopy())
			}
			d := schema.TestResourceDataRaw(t, Schema(), tc.config)
			result, err := util.ResourceConstruct(context.Background(), d, constructor)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			pool := result.(*kubeovnv1.IPPool)

			if !slices.Equal(pool.Spec.IPs, tc.expectedIPs) {
				t.Errorf("IPs: expected %v, got %v", tc.expectedIPs, pool.Spec.IPs)
			}
			if !slices.Equal(pool.Spec.Namespaces, tc.expectedNamespaces) {
				t.Errorf("Namespaces: expected %v, got %v", tc.expectedNamespaces, pool.Spec.Namespaces)
			}
		})
	}
}
