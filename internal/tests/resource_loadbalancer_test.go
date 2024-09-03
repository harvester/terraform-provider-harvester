package tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccLoadBalancer_invalid(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
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
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "harvester_virtualmachine" "test_vm" {
	name = "test-vm"
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

	depends_on = [
		harvester_virtualmachine.test_vm
	]

	listener {
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
