package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/harvester/harvester-network-controller/pkg/utils"
	"github.com/harvester/harvester/pkg/builder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	nadv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	kubeovnclient "github.com/kubeovn/kube-ovn/pkg/client/clientset/versioned"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

const (
	testAccKubeOVNDHCPNetwork = "test-acc-ovn-dhcp"
	testAccKubeOVNDHCPCIDR    = "10.241.0.0/24"
)

// testAccKubeOVNDHCPNetworkPreCheck creates an overlay network backed by a
// KubeOVN subnet with DHCP enabled, the case where the Harvester VM webhook
// replaces the bridge binding with managedtap. The test is skipped when
// KubeOVN is not installed.
func testAccKubeOVNDHCPNetworkPreCheck(ctx context.Context, t *testing.T) {
	testAccPreCheck(t)
	c, err := testAccProvider.Meta().(*config.Config).K8sClient()
	if err != nil {
		t.Fatal(err)
	}
	kubeOVN, err := kubeovnclient.NewForConfig(c.RestConfig)
	if err != nil {
		t.Fatal(err)
	}
	subnets := kubeOVN.KubeovnV1().Subnets()
	if _, err := subnets.List(ctx, metav1.ListOptions{Limit: 1}); err != nil {
		t.Skipf("KubeOVN is not available: %v", err)
	}

	provider := fmt.Sprintf("%s.default.ovn", testAccKubeOVNDHCPNetwork)
	nad := &nadv1.NetworkAttachmentDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testAccKubeOVNDHCPNetwork,
			Namespace: "default",
			Labels:    map[string]string{utils.KeyNetworkType: string(utils.OverlayNetwork)},
		},
		Spec: nadv1.NetworkAttachmentDefinitionSpec{
			Config: fmt.Sprintf(`{"cniVersion":"0.3.1","type":%q,"server_socket":"/run/openvswitch/kube-ovn-daemon.sock","provider":%q}`, utils.CNITypeKubeOVN, provider),
		},
	}
	nads := c.HarvesterClient.K8sCniCncfIoV1().NetworkAttachmentDefinitions("default")
	if _, err := nads.Create(ctx, nad, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := nads.Delete(context.Background(), testAccKubeOVNDHCPNetwork, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			t.Error(err)
		}
	})

	subnet := &kubeovnv1.Subnet{
		ObjectMeta: metav1.ObjectMeta{Name: testAccKubeOVNDHCPNetwork},
		Spec: kubeovnv1.SubnetSpec{
			Protocol:   kubeovnv1.ProtocolIPv4,
			CIDRBlock:  testAccKubeOVNDHCPCIDR,
			Provider:   provider,
			EnableDHCP: true,
		},
	}
	if _, err := subnets.Create(ctx, subnet, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		// The Harvester network webhook refuses to delete the network while
		// the subnet still uses it, so wait until the subnet is gone.
		ctx := context.Background()
		if err := subnets.Delete(ctx, testAccKubeOVNDHCPNetwork, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			t.Error(err)
			return
		}
		stateConf := getStateChangeConf(getResourceStateRefreshFunc(func() (interface{}, error) {
			return subnets.Get(ctx, testAccKubeOVNDHCPNetwork, metav1.GetOptions{})
		}))
		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			t.Error(err)
		}
	})
}

func testAccVirtualMachineKubeOVNConfig(memory string) string {
	return fmt.Sprintf(`
resource "harvester_virtualmachine" "test-acc-ovn-dhcp" {
	name      = "test-acc-ovn-dhcp"
	namespace = "default"

	cpu          = 1
	memory       = %q
	run_strategy = "Halted"
	machine_type = "q35"

	network_interface {
		name         = "nic-1"
		network_name = "default/%s"
	}

	disk {
		name       = "rootdisk"
		type       = "disk"
		bus        = "virtio"
		boot_order = 1

		container_image_name = %q
	}
}
`, memory, testAccKubeOVNDHCPNetwork, fedoraCloudContainer)
}

// testAccVirtualMachineInterfaceBinding checks the binding plugin set on the
// first interface of the VM.
func testAccVirtualMachineInterfaceBinding(ctx context.Context, n, binding string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vm, err := testAccGetVirtualMachine(ctx, s, n)
		if err != nil {
			return err
		}
		iface := vm.Spec.Template.Spec.Domain.Devices.Interfaces[0]
		if iface.Binding == nil || iface.Binding.Name != binding {
			return fmt.Errorf("interface binding is %+v, want %q", iface.Binding, binding)
		}
		return nil
	}
}

// TestAccVirtualMachine_kubeOVNDHCPNetwork checks that a VM on a KubeOVN
// subnet with DHCP can be managed: the Harvester webhook sets the managedtap
// binding, the provider reads the bridge type back, and the plan stays empty.
func TestAccVirtualMachine_kubeOVNDHCPNetwork(t *testing.T) {
	const resourceName = "harvester_virtualmachine.test-acc-ovn-dhcp"
	var (
		ctx      = context.Background()
		typeAttr = constants.FieldVirtualMachineNetworkInterface + ".0." + constants.FieldNetworkInterfaceType
	)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccKubeOVNDHCPNetworkPreCheck(ctx, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineKubeOVNConfig("1Gi"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, typeAttr, builder.NetworkInterfaceTypeBridge),
					testAccVirtualMachineInterfaceBinding(ctx, resourceName, constants.NetworkInterfaceBindingManagedTap),
				),
			},
			{
				Config: testAccVirtualMachineKubeOVNConfig("2Gi"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, typeAttr, builder.NetworkInterfaceTypeBridge),
					testAccVirtualMachineInterfaceBinding(ctx, resourceName, constants.NetworkInterfaceBindingManagedTap),
				),
			},
		},
	})
}
