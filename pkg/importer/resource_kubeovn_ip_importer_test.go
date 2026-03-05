package importer

import (
	"testing"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

func TestResourceKubeOVNIPStateGetter(t *testing.T) {
	testcases := []struct {
		name        string
		ip          *kubeovnv1.IP
		expectedID  string
		expectedPod string
		expectedIP  string
		expectedMAC string
	}{
		{
			name: "ip with all fields",
			ip: &kubeovnv1.IP{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pod.default",
				},
				Spec: kubeovnv1.IPSpec{
					PodName:     "test-pod",
					Namespace:   "default",
					Subnet:      "ovn-default",
					IPAddress:   "10.16.0.5",
					MacAddress:  "00:00:00:12:34:56",
					NodeName:    "node1",
					V4IPAddress: "10.16.0.5",
					V6IPAddress: "",
				},
			},
			expectedID:  helper.BuildID("", "test-pod.default"),
			expectedPod: "test-pod",
			expectedIP:  "10.16.0.5",
			expectedMAC: "00:00:00:12:34:56",
		},
		{
			name: "ip with nil labels",
			ip: &kubeovnv1.IP{
				ObjectMeta: metav1.ObjectMeta{
					Name: "minimal-ip",
				},
				Spec: kubeovnv1.IPSpec{
					PodName:   "pod1",
					Namespace: "kube-system",
					Subnet:    "join",
				},
			},
			expectedID:  helper.BuildID("", "minimal-ip"),
			expectedPod: "pod1",
			expectedIP:  "",
			expectedMAC: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getter, err := ResourceKubeOVNIPStateGetter(tc.ip)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if getter.ID != tc.expectedID {
				t.Errorf("ID: expected %q, got %q", tc.expectedID, getter.ID)
			}
			if getter.ResourceType != constants.ResourceTypeKubeOVNIP {
				t.Errorf("ResourceType: expected %q, got %q", constants.ResourceTypeKubeOVNIP, getter.ResourceType)
			}

			podName := getter.States[constants.FieldKubeOVNIPPodName].(string)
			if podName != tc.expectedPod {
				t.Errorf("PodName: expected %q, got %q", tc.expectedPod, podName)
			}

			ipAddr := getter.States[constants.FieldKubeOVNIPIPAddress].(string)
			if ipAddr != tc.expectedIP {
				t.Errorf("IPAddress: expected %q, got %q", tc.expectedIP, ipAddr)
			}

			mac := getter.States[constants.FieldKubeOVNIPMacAddress].(string)
			if mac != tc.expectedMAC {
				t.Errorf("MacAddress: expected %q, got %q", tc.expectedMAC, mac)
			}
		})
	}
}
