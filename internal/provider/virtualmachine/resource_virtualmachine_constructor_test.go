package virtualmachine

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

// TestUpdaterHostDevices checks that the update starts from no host device:
// the processors add back the configured ones, so a removed block detaches
// its device.
func TestUpdaterHostDevices(t *testing.T) {
	gpu := kubevirtv1.HostDevice{Name: "gpu", DeviceName: "nvidia.com/GA100_A30"}
	gmm := kubevirtv1.HostDevice{Name: "gmm", DeviceName: "intel.com/1911"}

	testcases := []struct {
		name        string
		hostDevices []kubevirtv1.HostDevice
	}{
		{
			name:        "one host device",
			hostDevices: []kubevirtv1.HostDevice{gmm},
		},
		{
			name:        "several host devices",
			hostDevices: []kubevirtv1.HostDevice{gpu, gmm},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			vm := &kubevirtv1.VirtualMachine{
				ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}},
				Spec:       kubevirtv1.VirtualMachineSpec{Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{Spec: kubevirtv1.VirtualMachineInstanceSpec{Domain: kubevirtv1.DomainSpec{Devices: kubevirtv1.Devices{HostDevices: tc.hostDevices}}}}},
			}
			Updater(nil, context.Background(), vm)
			if hostDevices := vm.Spec.Template.Spec.Domain.Devices.HostDevices; len(hostDevices) != 0 {
				t.Errorf("host devices before the processors run = %v, want none", hostDevices)
			}
		})
	}
}
