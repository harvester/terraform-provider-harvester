package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

const (
	testAccVirtualMachineName         = "test-acc-foo"
	testAccVirtualMachineResourceName = constants.ResourceTypeVirtualMachine + "." + testAccVirtualMachineName
	testAccVirtualMachineDescription  = "Terraform Harvester vm acceptance test"

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
)

func buildVirtualMachineConfig(name, description, memory string) string {
	return fmt.Sprintf(testAccVirtualMachineConfigTemplate, constants.ResourceTypeVirtualMachine, name,
		constants.FieldCommonName, name,
		constants.FieldCommonDescription, description,
		constants.FieldVirtualMachineMemory, memory)
}

func TestAccVirtualMachine_basic(t *testing.T) {
	var (
		vm  *kubevirtv1.VirtualMachine
		ctx = context.Background()
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

func TestAccVirtualMachine_labels(t *testing.T) {
	var (
		vm             *kubevirtv1.VirtualMachine
		ctx            = context.Background()
		expectedLabels = map[string]string{
			"tag.harvesterhci.io/Foobar": "barfoo",
		}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: `
resource "harvester_virtualmachine" "test-acc-labels" {
	name = "test-vm"
	namespace = "default"

	tags = {
		Foobar = "barfoo"
	}

	cpu = 1
	memory = "1Gi"

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
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVirtualMachineExists(ctx, "harvester_virtualmachine.test-acc-labels", vm),
					testAccVirtualMachineLabels(ctx, "harvester_virtualmachine.test-acc-labels", expectedLabels),
				),
			},
		},
	})
}

func testAccVirtualMachineExists(ctx context.Context, n string, vm *kubevirtv1.VirtualMachine) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		foundVM, err := testAccGetVirtualMachine(ctx, s, n)
		if err != nil {
			return err
		}
		vm = foundVM
		return nil
	}
}

func testAccVirtualMachineLabels(ctx context.Context, n string, labels map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vm, err := testAccGetVirtualMachine(ctx, s, n)
		if err != nil {
			return err
		}

		for key := range labels {
			val, ok := vm.Labels[key]
			if !ok {
				return fmt.Errorf("Label %s not found", key)
			}

			if val != labels[key] {
				return fmt.Errorf("Label %s contains unexpected value: %s", key, val)
			}
		}
		return nil
	}
}

func testAccGetVirtualMachine(ctx context.Context, state *terraform.State, resourceName string) (*kubevirtv1.VirtualMachine, error) {
	resource, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("Resource not found: %s", resourceName)
	}
	if resource.Primary.ID == "" {
		return nil, fmt.Errorf("Resource ID not set: %s", resourceName)
	}

	id := resource.Primary.ID
	c := testAccProvider.Meta().(*client.Client)

	namespace, name, err := helper.IDParts(id)
	if err != nil {
		return nil, err
	}

	return c.HarvesterClient.
		KubevirtV1().
		VirtualMachines(namespace).
		Get(ctx, name, metav1.GetOptions{})
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
