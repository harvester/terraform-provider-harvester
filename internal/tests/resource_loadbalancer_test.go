package tests

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAccLoadBalancer_invalid(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLoadBalancerDestroy(context.Background()),
		Steps: []resource.TestStep{
			{
				Config: `
resource "harvester_loadbalancer" "test_loadbalancer" {
}
`,
				ExpectError: regexp.MustCompile(`The argument "name" is required, but no definition was found.`),
			},
			{
				Config: `
resource "harvester_loadbalancer" "test_loadbalancer" {
	name = "test-loadbalancer"
}
`,
				ExpectError: regexp.MustCompile(`The argument "listener" is required, but no definition was found.`),
			},
		},
	})
}

func TestAccLoadBalancer_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLoadBalancerDestroy(context.Background()),
		Steps: []resource.TestStep{
			{
				Config: `
resource "harvester_virtualmachine" "test_vm" {
	name = "test-vm-for-lb"
	namespace = "default"

	tags = {
		app = "testlb"
	}

	cpu = 1
	memory = "1Gi"
	machine_type = "q35"
	run_strategy = "RerunOnFailure"

	network_interface {
		name = "default"
	}

	disk {
		name = "rootdisk"
		type = "disk"
		bus = "virtio"
		boot_order = 1

		container_image_name = "kubevirt/fedora-cloud-container-disk-demo:v0.35.0"
	}
}

resource "harvester_loadbalancer" "test_loadbalancer" {
	name = "test-loadbalancer"
	namespace = "default"

	depends_on = [
		harvester_virtualmachine.test_vm
	]

	ipam = "dhcp"

	listener {
		name = "https"
		port = 443
		protocol = "tcp"
		backend_port = 8080
	}

	backend_selector {
		key = "tag.harvesterhci.io/app"
		values = [ "testlb" ]
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("harvester_loadbalancer.test_loadbalancer", "name", "test-loadbalancer"),
					resource.TestCheckResourceAttr("harvester_loadbalancer.test_loadbalancer", "listener.#", "1"),
					resource.TestCheckResourceAttr("harvester_loadbalancer.test_loadbalancer", "listener.0.port", "443"),
					resource.TestCheckResourceAttr("harvester_loadbalancer.test_loadbalancer", "listener.0.protocol", "tcp"),
					resource.TestCheckResourceAttr("harvester_loadbalancer.test_loadbalancer", "listener.0.backend_port", "8080"),
				),
			},
		},
	})
}

func testAccCheckLoadBalancerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			c, err := testAccProvider.Meta().(*config.Config).K8sClient()
			if err != nil {
				return err
			}

			namespace, name, err := helper.IDParts(rs.Primary.ID)
			if err != nil {
				return err
			}

			switch t := rs.Type; t {
			case constants.ResourceTypeLoadBalancer:
				lbStateRefreshFunc := getResourceStateRefreshFunc(func() (interface{}, error) {
					return c.HarvesterLoadbalancerClient.
						LoadbalancerV1beta1().
						LoadBalancers(namespace).
						Get(ctx, name, metav1.GetOptions{})
				})
				stateConf := getStateChangeConf(lbStateRefreshFunc)
				if _, err = stateConf.WaitForStateContext(ctx); err != nil {
					return fmt.Errorf("[ERROR] waiting for loadbalancer (%s) to be removed: %s", rs.Primary.ID, err)
				}
			case constants.ResourceTypeVirtualMachine:
				vmStateRefreshFunc := getResourceStateRefreshFunc(func() (interface{}, error) {
					return c.HarvesterClient.
						KubevirtV1().
						VirtualMachines(namespace).
						Get(ctx, name, metav1.GetOptions{})
				})
				stateConf := getStateChangeConf(vmStateRefreshFunc)
				if _, err = stateConf.WaitForStateContext(ctx); err != nil {
					return fmt.Errorf("[ERROR] waiting for virtual machine (%s) to be removed: %s", rs.Primary.ID, err)
				}
			}
		}
		return nil
	}
}
