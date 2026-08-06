package importer

import (
	"testing"

	assert "github.com/stretchr/testify/assert"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func TestStateRunningNetworkInterfaceNoIP(t *testing.T) {
	// Ensure that of `wait_for_lease` is set on a network interface, the state isn't reported as
	// `ready` as long as there is no IP address
	oldInstanceUID := "oldUid"
	networkInterfaces := []map[string]any{
		{
			constants.FieldNetworkInterfaceWaitForLease: true,
		},
	}
	vmi := kubevirtv1.VirtualMachineInstance{
		ObjectMeta: metav1.ObjectMeta{
			UID: "newUid",
		},
		Status: kubevirtv1.VirtualMachineInstanceStatus{
			Phase: "Running",
		},
	}

	vmImporter := NewVMImporter(nil, &vmi)
	state := vmImporter.State(networkInterfaces, oldInstanceUID)
	assert.Equal(t, state, constants.StateVirtualMachineRunning, "IP address not set results in state running")

	networkInterfaces[0][constants.FieldNetworkInterfaceIPAddress] = ""
	state = vmImporter.State(networkInterfaces, oldInstanceUID)
	assert.Equal(t, state, constants.StateVirtualMachineRunning, "IP address set to empty string results in state running")

	networkInterfaces[0][constants.FieldNetworkInterfaceIPAddress] = "123.123.123.123"
	state = vmImporter.State(networkInterfaces, oldInstanceUID)
	assert.Equal(t, state, constants.StateCommonReady, "IP address set to non-empty string results in state ready")
}
