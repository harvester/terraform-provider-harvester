package tests

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

const (
	testAccVirtualMachineDescription = "Terraform Harvester vm acceptance test"

	testAccVirtualMachineMemory       = "1Gi"
	testAccVirtualMachineMemoryUpdate = "2Gi"

	testAccVirtualMachineConfigTemplate = `
resource %s "%s" {
	%s = "%s"
	%s = "%s"

  cpu = 1
  %s = "%s"

  run_strategy = "RerunOnFailure"
  machine_type = "q35"

  network_interface {
    name         = "default"
  }

  disk {
    name       = "rootdisk"
    type       = "disk"
    bus        = "virtio"
    boot_order = 1

    container_image_name = "kubevirt/fedora-cloud-container-disk-demo:v0.35.0"
  }
}
`
	testAccInputBlockTemplate = `
  input {
		name = "%s"
		type = "%s"
		bus = "%s"
	}
`
)

func addInputBlockConfig(name, inputType, bus, vmConfig string) string {
	inputBlock := fmt.Sprintf(testAccInputBlockTemplate, name, inputType, bus)
	return vmConfig[:(len(vmConfig)-3)] + inputBlock + vmConfig[(len(vmConfig)-3):]
}

func buildVirtualMachineConfig(name, description, memory string) string {
	return fmt.Sprintf(testAccVirtualMachineConfigTemplate,
		constants.ResourceTypeVirtualMachine, name,
		constants.FieldCommonName, name,
		constants.FieldCommonDescription, description,
		constants.FieldVirtualMachineMemory, memory)
}

func TestAccVirtualMachine_basic(t *testing.T) {
	var (
		testAccVirtualMachineName         = "test-acc-basic-" + uuid.New().String()[:6]
		testAccVirtualMachineResourceName = constants.ResourceTypeVirtualMachine + "." + testAccVirtualMachineName
		vm                                *kubevirtv1.VirtualMachine
		ctx                               = context.Background()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: buildVirtualMachineConfig(testAccVirtualMachineName, testAccVirtualMachineDescription, testAccVirtualMachineMemory),
				Check: resource.ComposeTestCheckFunc(
					testAccVirtualMachineExists(ctx, testAccVirtualMachineResourceName, vm),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldCommonName, testAccVirtualMachineName),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldCommonDescription, testAccVirtualMachineDescription),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineMemory, testAccVirtualMachineMemory),
				),
			},
			{
				Config: buildVirtualMachineConfig(testAccVirtualMachineName, testAccVirtualMachineDescription, testAccVirtualMachineMemoryUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccVirtualMachineExists(ctx, testAccVirtualMachineResourceName, vm),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldCommonName, testAccVirtualMachineName),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldCommonDescription, testAccVirtualMachineDescription),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineMemory, testAccVirtualMachineMemoryUpdate),
				),
			},
		},
	})
}

func TestAccVirtualMachine_input(t *testing.T) {
	var (
		testAccVirtualMachineName         = "test-acc-input-" + uuid.New().String()[:6]
		testAccVirtualMachineResourceName = constants.ResourceTypeVirtualMachine + "." + testAccVirtualMachineName
		vm                                *kubevirtv1.VirtualMachine
		ctx                               = context.Background()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: addInputBlockConfig(
					"tablet", "tablet", "usb",
					buildVirtualMachineConfig(
						testAccVirtualMachineName,
						testAccVirtualMachineDescription,
						testAccVirtualMachineMemoryUpdate,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccVirtualMachineExists(ctx, testAccVirtualMachineResourceName, vm),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineInput+".#", "1"),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineInput+".0.name", "tablet"),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineInput+".0.type", "tablet"),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineInput+".0.bus", "usb"),
				),
			},
		},
	})
}

func TestAccVirtualMachine_disk_size(t *testing.T) {
	var (
		testAccImageName         = "test-acc-image-leap-" + uuid.New().String()[:6]
		testAccImageResourceName = constants.ResourceTypeImage + "." + testAccImageName
		ctx                      = context.Background()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource harvester_image "%s" {
  name = "leap-15.6"
	namespace = "default"
	display_name = "openSUSE-Leap-15.6"
	source_type = "download"
	url = "https://download.opensuse.org/repositories/Cloud:/Images:/Leap_15.6/images/openSUSE-Leap-15.6.x86_64-NoCloud.qcow2"
}
`,
					testAccImageName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testAccImageResourceName, constants.FieldCommonName, "leap-15.6"),
					resource.TestCheckResourceAttr(testAccImageResourceName, constants.FieldCommonNamespace, "default"),
				),
			},
			{
				Config: `
resource harvester_virtualmachine "disk_test" {
	name = "disk-test"

  cpu = 1
  memory = "1Gi"

  run_strategy = "RerunOnFailure"
  machine_type = "q35"

  network_interface {
    name         = "default"
  }

  disk {
    name       = "cdrom-disk"
    type       = "cd-rom"
    bus        = "sata"
    boot_order = 1
		size       = "foobar"
    image      = "default/leap-15.6"
  }
}
`,
				ExpectError: regexp.MustCompile(".*is not a parsable quantity.*"),
				Check:       resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func testAccVirtualMachineExists(ctx context.Context, n string, vm *kubevirtv1.VirtualMachine) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Resource %s not found. ", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource %s ID not set. ", n)
		}

		id := rs.Primary.ID
		c := testAccProvider.Meta().(*client.Client)

		namespace, name, err := helper.IDParts(id)
		if err != nil {
			return err
		}
		foundVM, err := c.HarvesterClient.KubevirtV1().VirtualMachines(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		vm = foundVM
		return nil
	}
}

func testAccCheckVirtualMachineDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != constants.ResourceTypeVirtualMachine {
				continue
			}

			c := testAccProvider.Meta().(*client.Client)
			namespace, name, err := helper.IDParts(rs.Primary.ID)
			if err != nil {
				return err
			}

			virtualMachineStateRefreshFunc := getResourceStateRefreshFunc(func() (interface{}, error) {
				return c.HarvesterClient.KubevirtV1().VirtualMachines(namespace).Get(ctx, name, metav1.GetOptions{})
			})
			stateConf := getStateChangeConf(virtualMachineStateRefreshFunc)
			if _, err = stateConf.WaitForStateContext(ctx); err != nil {
				return fmt.Errorf(
					"[ERROR] waiting for virtual machine (%s) to be removed: %s", rs.Primary.ID, err)
			}
		}
		return nil
	}
}
