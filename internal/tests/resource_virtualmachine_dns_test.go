package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	corev1 "k8s.io/api/core/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func TestAccVirtualMachine_dns_config(t *testing.T) {
	var (
		testAccName         = "test-acc-dns-" + uuid.New().String()[:6]
		testAccResourceName = constants.ResourceTypeVirtualMachine + "." + testAccName
		vm                  = &kubevirtv1.VirtualMachine{}
		ctx                 = context.Background()
	)

	vmConfig := fmt.Sprintf(`
resource %s "%s" {
	name = "%s"

	cpu    = 1
	memory = "1Gi"

	run_strategy = "RerunOnFailure"
	machine_type = "q35"

	dns_policy = "None"

	dns_config {
		nameservers = ["8.8.8.8", "8.8.4.4"]
		searches    = ["example.com", "test.local"]

		options {
			name  = "ndots"
			value = "5"
		}
		options {
			name = "single-request-reopen"
		}
	}

	network_interface {
		name = "default"
	}

	disk {
		name               = "rootdisk"
		type               = "disk"
		bus                = "virtio"
		boot_order         = 1
		container_image_name = "%s"
	}
}
`, constants.ResourceTypeVirtualMachine, testAccName, testAccName, fedoraCloudContainer)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: vmConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccVirtualMachineExists(ctx, testAccResourceName, vm),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineDNSPolicy, "None"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineDNSConfig+".#", "1"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineDNSConfig+".0.nameservers.#", "2"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineDNSConfig+".0.nameservers.0", "8.8.8.8"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineDNSConfig+".0.nameservers.1", "8.8.4.4"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineDNSConfig+".0.searches.#", "2"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineDNSConfig+".0.options.#", "2"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineDNSConfig+".0.options.0.name", "ndots"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineDNSConfig+".0.options.0.value", "5"),
					testAccCheckVMDNSConfig(ctx, testAccResourceName),
				),
			},
		},
	})
}

// testAccCheckVMDNSConfig verifies the K8s VM object has the expected DNS configuration.
func testAccCheckVMDNSConfig(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vm, err := testAccGetVirtualMachine(ctx, s, n)
		if err != nil {
			return err
		}
		if vm.Spec.Template == nil {
			return fmt.Errorf("VM template is nil")
		}
		if vm.Spec.Template.Spec.DNSPolicy != corev1.DNSNone {
			return fmt.Errorf("expected DNS policy None, got %s", vm.Spec.Template.Spec.DNSPolicy)
		}
		dnsConfig := vm.Spec.Template.Spec.DNSConfig
		if dnsConfig == nil {
			return fmt.Errorf("DNS config is nil")
		}
		if len(dnsConfig.Nameservers) != 2 || dnsConfig.Nameservers[0] != "8.8.8.8" {
			return fmt.Errorf("unexpected nameservers: %v", dnsConfig.Nameservers)
		}
		if len(dnsConfig.Searches) != 2 || dnsConfig.Searches[0] != "example.com" {
			return fmt.Errorf("unexpected searches: %v", dnsConfig.Searches)
		}
		if len(dnsConfig.Options) != 2 {
			return fmt.Errorf("expected 2 DNS options, got %d", len(dnsConfig.Options))
		}
		return nil
	}
}
