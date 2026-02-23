package tests

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

const (
	defaultVMDesciption  = "Terraform Harvester vm acceptance test"
	defaultVMMemory      = "1Gi"
	fedoraCloudContainer = "kubevirt/fedora-cloud-container-disk-demo:v0.35.0"
)

type VMResourceBuilder struct {
	name                          string
	description                   string
	cpu                           int
	memory                        string // e.g. "1Gi"
	cpuPinning                    bool
	isolateEmulatorThread         bool
	runStrategy                   string
	machineType                   string
	evictionStrategy              string
	terminationGracePeriodSeconds *int
	osType                        string
	networkConfig                 *NetworkConfig
	diskConfig                    *DiskConfig
	inputConfig                   *InputDeviceConfig
}

type DiskConfig struct {
	Name               string
	Type               string
	Bus                string
	BootOrder          int
	ContainerImageName string
}

type NetworkConfig struct {
	Name      string
	BootOrder int
}

type InputDeviceConfig struct {
	Name string
	Type string
	Bus  string
}

func NewVMResourceBuilder(resourceName string) *VMResourceBuilder {
	return &VMResourceBuilder{
		name:                  resourceName,
		description:           defaultVMDesciption,
		cpu:                   1,
		cpuPinning:            false,
		isolateEmulatorThread: false,
		memory:                defaultVMMemory,
		runStrategy:           "RerunOnFailure",
		machineType:           "q35",
		networkConfig: &NetworkConfig{
			Name:      "default",
			BootOrder: 0,
		},
		diskConfig: &DiskConfig{
			Name:               "rootdisk",
			Type:               "disk",
			Bus:                "virtio",
			BootOrder:          1,
			ContainerImageName: fedoraCloudContainer,
		},
	}
}

func (b *VMResourceBuilder) SetMemory(memory string) *VMResourceBuilder {
	b.memory = memory
	return b
}

func (b *VMResourceBuilder) SetCPUPinning(cpuPinning bool) *VMResourceBuilder {
	b.cpuPinning = cpuPinning
	return b
}

func (b *VMResourceBuilder) SetIsolateEmulatorThread(isolateEmulatorThread bool) *VMResourceBuilder {
	b.isolateEmulatorThread = isolateEmulatorThread
	return b
}

func (b *VMResourceBuilder) SetInputDeviceConfig(name, inputType, bus string) *VMResourceBuilder {
	b.inputConfig = &InputDeviceConfig{
		Name: name,
		Type: inputType,
		Bus:  bus,
	}
	return b
}

func (b *VMResourceBuilder) SetEvictionStrategy(strategy string) *VMResourceBuilder {
	b.evictionStrategy = strategy
	return b
}

func (b *VMResourceBuilder) SetTerminationGracePeriodSeconds(seconds int) *VMResourceBuilder {
	b.terminationGracePeriodSeconds = &seconds
	return b
}

func (b *VMResourceBuilder) SetOSType(osType string) *VMResourceBuilder {
	b.osType = osType
	return b
}

func (b *VMResourceBuilder) SetNetworkConfig(name string, bootOrder int) *VMResourceBuilder {
	b.networkConfig = &NetworkConfig{
		Name:      name,
		BootOrder: bootOrder,
	}
	return b
}

func (b *VMResourceBuilder) SetDiskConfig(name, bus, image string, bootOrder int) *VMResourceBuilder {
	b.diskConfig = &DiskConfig{
		Name:               name,
		Type:               "disk",
		Bus:                bus,
		BootOrder:          bootOrder,
		ContainerImageName: image,
	}
	return b
}

// Build generates the terraform resource string.
func (b *VMResourceBuilder) Build() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("resource %s \"%s\" {\n", constants.ResourceTypeVirtualMachine, b.name))

	sb.WriteString(fmt.Sprintf("\t%s = \"%s\"\n", constants.FieldCommonName, b.name))
	sb.WriteString(fmt.Sprintf("\t%s = \"%s\"\n", constants.FieldCommonDescription, b.description))

	sb.WriteString(fmt.Sprintf("\t%s = %d\n", constants.FieldVirtualMachineCPU, b.cpu))
	sb.WriteString(fmt.Sprintf("\t%s = \"%s\"\n", constants.FieldVirtualMachineMemory, b.memory))
	sb.WriteString(fmt.Sprintf("\t%s = %s\n", constants.FieldVirtualMachineCPUPinning, strconv.FormatBool(b.cpuPinning)))
	sb.WriteString(fmt.Sprintf("\t%s = %s\n", constants.FieldVirtualMachineIsolateEmulatorThread, strconv.FormatBool(b.isolateEmulatorThread)))
	sb.WriteString(fmt.Sprintf("\t%s = \"%s\"\n", constants.FieldVirtualMachineRunStrategy, b.runStrategy))
	sb.WriteString(fmt.Sprintf("\t%s = \"%s\"\n", constants.FieldVirtualMachineMachineType, b.machineType))

	if b.evictionStrategy != "" {
		sb.WriteString(fmt.Sprintf("\t%s = \"%s\"\n", constants.FieldVirtualMachineEvictionStrategy, b.evictionStrategy))
	}
	if b.terminationGracePeriodSeconds != nil {
		sb.WriteString(fmt.Sprintf("\t%s = %d\n", constants.FieldVirtualMachineTerminationGracePeriodSeconds, *b.terminationGracePeriodSeconds))
	}
	if b.osType != "" {
		sb.WriteString(fmt.Sprintf("\t%s = \"%s\"\n", constants.FieldVirtualMachineOSType, b.osType))
	}

	if b.networkConfig != nil {
		sb.WriteString(fmt.Sprintf("\t%s {\n", constants.FieldVirtualMachineNetworkInterface))
		sb.WriteString(fmt.Sprintf("\t\t%s = \"%s\"\n", constants.FieldNetworkInterfaceName, b.networkConfig.Name))
		sb.WriteString(fmt.Sprintf("\t\t%s = %d\n", constants.FieldNetworkInterfaceBootOrder, b.networkConfig.BootOrder))
		sb.WriteString("\t}\n")
	}

	if b.diskConfig != nil {
		sb.WriteString(fmt.Sprintf("\t%s {\n", constants.FieldVirtualMachineDisk))
		sb.WriteString(fmt.Sprintf("\t\t%s = \"%s\"\n", constants.FieldDiskName, b.diskConfig.Name))
		sb.WriteString(fmt.Sprintf("\t\t%s = \"%s\"\n", constants.FieldDiskType, b.diskConfig.Type))
		sb.WriteString(fmt.Sprintf("\t\t%s = \"%s\"\n", constants.FieldDiskBus, b.diskConfig.Bus))
		sb.WriteString(fmt.Sprintf("\t\t%s = %d\n", constants.FieldDiskBootOrder, b.diskConfig.BootOrder))
		sb.WriteString(fmt.Sprintf("\t\t%s = \"%s\"\n", constants.FieldDiskContainerImageName, b.diskConfig.ContainerImageName))
		sb.WriteString("\t}\n")
	}

	if b.inputConfig != nil {
		sb.WriteString(fmt.Sprintf("\t%s {\n", constants.FieldVirtualMachineInput))
		sb.WriteString(fmt.Sprintf("\t\t%s = \"%s\"\n", constants.FieldInputName, b.inputConfig.Name))
		sb.WriteString(fmt.Sprintf("\t\t%s = \"%s\"\n", constants.FieldInputType, b.inputConfig.Type))
		sb.WriteString(fmt.Sprintf("\t\t%s = \"%s\"\n", constants.FieldInputBus, b.inputConfig.Bus))
		sb.WriteString("\t}\n")
	}

	sb.WriteString("}\n")
	return sb.String()
}

func TestAccVirtualMachine_basic(t *testing.T) {
	var (
		testAccVirtualMachineName         = "test-acc-basic-" + uuid.New().String()[:6]
		testAccVirtualMachineResourceName = constants.ResourceTypeVirtualMachine + "." + testAccVirtualMachineName
		testAccVirtualMachineMemoryUpdate = "2Gi"
		vm                                = &kubevirtv1.VirtualMachine{}
		ctx                               = context.Background()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: NewVMResourceBuilder(testAccVirtualMachineName).Build(),
				Check: resource.ComposeTestCheckFunc(
					testAccVirtualMachineExists(ctx, testAccVirtualMachineResourceName, vm),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldCommonName, testAccVirtualMachineName),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldCommonDescription, defaultVMDesciption),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineMemory, defaultVMMemory),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineCPUPinning, "false"),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineIsolateEmulatorThread, "false"),
				),
			},
			{
				Config: NewVMResourceBuilder(testAccVirtualMachineName).SetMemory(testAccVirtualMachineMemoryUpdate).Build(),
				Check: resource.ComposeTestCheckFunc(
					testAccVirtualMachineExists(ctx, testAccVirtualMachineResourceName, vm),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldCommonName, testAccVirtualMachineName),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldCommonDescription, defaultVMDesciption),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineMemory, testAccVirtualMachineMemoryUpdate),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineCPUPinning, "false"),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineIsolateEmulatorThread, "false"),
				),
			},
		},
	})
}

func TestAccVirtualMachine_network_boot(t *testing.T) {
	var (
		testAccVirtualMachineName         = "test-acc-netboot-" + uuid.New().String()[:6]
		testAccVirtualMachineResourceName = constants.ResourceTypeVirtualMachine + "." + testAccVirtualMachineName
		vm                                = &kubevirtv1.VirtualMachine{}
		ctx                               = context.Background()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: NewVMResourceBuilder(testAccVirtualMachineName).
					SetNetworkConfig("default", 1).
					SetDiskConfig("rootdisk", "virtio", fedoraCloudContainer, 2).
					Build(),
				Check: resource.ComposeTestCheckFunc(
					testAccVirtualMachineExists(ctx, testAccVirtualMachineResourceName, vm),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldCommonName, testAccVirtualMachineName),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldCommonDescription, defaultVMDesciption),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineMemory, defaultVMMemory),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineCPUPinning, "false"),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineIsolateEmulatorThread, "false"),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineNetworkInterface+".#", "1"),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineNetworkInterface+".0.name", "default"),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineNetworkInterface+".0.boot_order", "1"),
				),
			},
		},
	})
}

func TestAccVirtualMachine_cpu_pinning(t *testing.T) {
	var (
		testAccVirtualMachineName         = "test-acc-cpu-pinning" + uuid.New().String()[:6]
		testAccVirtualMachineNamespace    = "default"
		testAccVirtualMachineResourceName = constants.ResourceTypeVirtualMachine + "." + testAccVirtualMachineName
		vm                                = &kubevirtv1.VirtualMachine{}
		ctx                               = context.Background()
	)

	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping test: TF_ACC is not set")
	}

	// enableCPUManager and disableCPUManager are placed outside acc test because acc tests lack beforeAll/afterAll support.
	// To minimize delays, CPUManager is enable/disable only once rather than before/after each test step.
	testAccPreCheck(t)

	c, err := testAccProvider.Meta().(*config.Config).K8sClient()
	if err != nil {
		t.Fatal("failed to find any node")
	}
	ensureNodeManagerEnabled(ctx, t, c)

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: NewVMResourceBuilder(testAccVirtualMachineName).SetCPUPinning(true).Build(),
				Check: resource.ComposeTestCheckFunc(
					testAccVirtualMachineExists(ctx, testAccVirtualMachineResourceName, vm),
					func(s *terraform.State) error {
						vm, err = c.HarvesterClient.KubevirtV1().VirtualMachines(testAccVirtualMachineNamespace).Get(ctx, testAccVirtualMachineName, metav1.GetOptions{})
						if err != nil {
							return err
						}
						if vm.Spec.Template == nil || vm.Spec.Template.Spec.Domain.CPU == nil || !vm.Spec.Template.Spec.Domain.CPU.DedicatedCPUPlacement {
							return errors.New("DedicatedCPUPlacement should be true")
						}
						if vm.Spec.Template == nil || vm.Spec.Template.Spec.Domain.CPU == nil || vm.Spec.Template.Spec.Domain.CPU.IsolateEmulatorThread {
							return errors.New("IsolateEmulatorThread should be false")
						}
						return nil
					},
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldCommonName, testAccVirtualMachineName),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldCommonDescription, defaultVMDesciption),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineMemory, defaultVMMemory),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineCPUPinning, "true"),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineIsolateEmulatorThread, "false"),
				),
			},
			{
				Config: NewVMResourceBuilder(testAccVirtualMachineName).SetCPUPinning(true).SetIsolateEmulatorThread(true).Build(),
				Check: resource.ComposeTestCheckFunc(
					testAccVirtualMachineExists(ctx, testAccVirtualMachineResourceName, vm),
					func(s *terraform.State) error {
						vm, err = c.HarvesterClient.KubevirtV1().VirtualMachines(testAccVirtualMachineNamespace).Get(ctx, testAccVirtualMachineName, metav1.GetOptions{})
						if err != nil {
							return err
						}
						if vm.Spec.Template == nil || vm.Spec.Template.Spec.Domain.CPU == nil || !vm.Spec.Template.Spec.Domain.CPU.DedicatedCPUPlacement {
							return errors.New("DedicatedCPUPlacement should be true")
						}
						if vm.Spec.Template == nil || vm.Spec.Template.Spec.Domain.CPU == nil || !vm.Spec.Template.Spec.Domain.CPU.IsolateEmulatorThread {
							return errors.New("IsolateEmulatorThread should be true")
						}
						return nil
					},
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldCommonName, testAccVirtualMachineName),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldCommonDescription, defaultVMDesciption),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineMemory, defaultVMMemory),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineCPUPinning, "true"),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineIsolateEmulatorThread, "true"),
				),
			},
		},
	})
}

func ensureNodeManagerEnabled(ctx context.Context, t *testing.T, c *client.Client) {
	nodes, err := c.KubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes.Items) == 0 {
		t.Fatal("failed to find any node")
	}
	node := nodes.Items[0]
	nodeName := node.Name
	enabled, ok := node.Labels["cpumanager"]
	if !ok || enabled != "true" {
		t.Log("enable cpu manager on node " + nodeName)
		enableCPUManager(t, ctx, nodeName)
	}
}

func TestAccVirtualMachine_input(t *testing.T) {
	var (
		testAccVirtualMachineName         = "test-acc-input-" + uuid.New().String()[:6]
		testAccVirtualMachineResourceName = constants.ResourceTypeVirtualMachine + "." + testAccVirtualMachineName
		vm                                = &kubevirtv1.VirtualMachine{}
		ctx                               = context.Background()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: NewVMResourceBuilder(testAccVirtualMachineName).SetInputDeviceConfig("tablet", "tablet", "usb").Build(),
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
	storage_class_name = "harvester-longhorn"
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
	name = "test-vm-for-labels"
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

func TestAccVirtualMachine_hotplug_cdrom_volume(t *testing.T) {
	var (
		testAccImageName                  = "test-acc-hp-cdrom-img"
		testAccVirtualMachineName         = "test-acc-hp-cdrom"
		testAccVirtualMachineNamespace    = "default"
		testAccImageResourceName          = constants.ResourceTypeImage + "." + testAccImageName
		testAccVirtualMachineResourceName = constants.ResourceTypeVirtualMachine + "." + testAccVirtualMachineName
		vm                                = &kubevirtv1.VirtualMachine{}
		vmiUid                            types.UID
		ctx                               = context.Background()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "harvester_virtualmachine" "%s" {
	name = "%s"
	namespace = "%s"

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

  disk {
    name       = "cd1"
    type       = "cd-rom"
    bus        = "sata"
    boot_order = 2
  }
}
`,
					testAccVirtualMachineName, testAccVirtualMachineName, testAccVirtualMachineNamespace,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVirtualMachineExists(ctx, testAccVirtualMachineResourceName, vm),
					testAccCheckCdRomSpec(ctx, testAccVirtualMachineNamespace, testAccVirtualMachineName, 2, 1),
					func(s *terraform.State) error {
						c, err := testAccProvider.Meta().(*config.Config).K8sClient()
						if err != nil {
							return err
						}
						vmi, err := c.HarvesterClient.KubevirtV1().VirtualMachineInstances(testAccVirtualMachineNamespace).Get(ctx, testAccVirtualMachineName, metav1.GetOptions{})
						if err != nil {
							return err
						}
						vmiUid = vmi.UID
						return nil
					},
				),
			},
			{
				Config: fmt.Sprintf(`
resource harvester_image "%s" {
  name = "%s"
	namespace = "%s"
	display_name = "%s"
	source_type = "download"
	url = "https://distro.ibiblio.org/tinycorelinux/16.x/x86/release/TinyCore-current.iso"
	storage_class_name = "harvester-longhorn"
}

resource "harvester_virtualmachine" "%s" {
	name = "%s"
	namespace = "%s"

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

  disk {
    name       = "cd1"
    type       = "cd-rom"
    bus        = "sata"
    boot_order = 2

    size        = "1Gi"
    hot_plug    = true
    image       = "%s/%s"
    auto_delete = true
  }
}
`,
					testAccImageName, testAccImageName, testAccVirtualMachineNamespace, testAccImageName,
					testAccVirtualMachineName, testAccVirtualMachineName, testAccVirtualMachineNamespace,
					testAccVirtualMachineNamespace, testAccImageName,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccImageResourceName, constants.FieldCommonName, "test-acc-hp-cdrom-img"),
					resource.TestCheckResourceAttr(testAccImageResourceName, constants.FieldCommonNamespace, "default"),
					testAccVirtualMachineExists(ctx, testAccVirtualMachineResourceName, vm),
					testAccCheckCdRomSpec(ctx, testAccVirtualMachineNamespace, testAccVirtualMachineName, 2, 2),
					testAccCheckVmiUid(ctx, testAccVirtualMachineNamespace, testAccVirtualMachineName, &vmiUid),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "harvester_virtualmachine" "%s" {
	name = "%s"
	namespace = "%s"

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

  disk {
    name       = "cd1"
    type       = "cd-rom"
    bus        = "sata"
    boot_order = 2
  }
}
`,
					testAccVirtualMachineName, testAccVirtualMachineName, testAccVirtualMachineNamespace,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVirtualMachineExists(ctx, testAccVirtualMachineResourceName, vm),
					testAccCheckCdRomSpec(ctx, testAccVirtualMachineNamespace, testAccVirtualMachineName, 2, 1),
					testAccCheckVmiUid(ctx, testAccVirtualMachineNamespace, testAccVirtualMachineName, &vmiUid),
				),
			},
		},
	})
}

func TestAccVirtualMachine_runtime_options(t *testing.T) {
	var (
		testAccVirtualMachineName         = "test-acc-runtime-" + uuid.New().String()[:6]
		testAccVirtualMachineResourceName = constants.ResourceTypeVirtualMachine + "." + testAccVirtualMachineName
		vm                                = &kubevirtv1.VirtualMachine{}
		ctx                               = context.Background()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: NewVMResourceBuilder(testAccVirtualMachineName).
					SetEvictionStrategy("LiveMigrateIfPossible").
					SetTerminationGracePeriodSeconds(60).
					SetOSType("linux").
					Build(),
				Check: resource.ComposeTestCheckFunc(
					testAccVirtualMachineExists(ctx, testAccVirtualMachineResourceName, vm),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineEvictionStrategy, "LiveMigrateIfPossible"),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineTerminationGracePeriodSeconds, "60"),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineOSType, "linux"),
				),
			},
			{
				Config: NewVMResourceBuilder(testAccVirtualMachineName).
					SetEvictionStrategy("None").
					SetTerminationGracePeriodSeconds(60).
					SetOSType("linux").
					Build(),
				Check: resource.ComposeTestCheckFunc(
					testAccVirtualMachineExists(ctx, testAccVirtualMachineResourceName, vm),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineEvictionStrategy, "None"),
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

func testAccCheckCdRomSpec(ctx context.Context, vmNamespace, vmName string, expectedDisksCnt, expectedVolumeCnts int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c, err := testAccProvider.Meta().(*config.Config).K8sClient()
		if err != nil {
			return err
		}
		vm, err := c.HarvesterClient.KubevirtV1().VirtualMachines(vmNamespace).Get(ctx, vmName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		disksCnt := len(vm.Spec.Template.Spec.Domain.Devices.Disks)
		if disksCnt != expectedDisksCnt {
			return fmt.Errorf("Should have %d disk devices but got %d", expectedDisksCnt, disksCnt)
		}
		volumeCnts := len(vm.Spec.Template.Spec.Volumes)
		if volumeCnts != expectedVolumeCnts {
			return fmt.Errorf("Should have %d volumes but got %d", expectedVolumeCnts, volumeCnts)
		}

		return nil
	}
}

func testAccCheckVmiUid(ctx context.Context, vmNamespace, vmName string, vmiUid *types.UID) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c, err := testAccProvider.Meta().(*config.Config).K8sClient()
		if err != nil {
			return err
		}
		vmi, err := c.HarvesterClient.KubevirtV1().VirtualMachineInstances(vmNamespace).Get(ctx, vmName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if vmi.UID != *vmiUid {
			return fmt.Errorf("Shouldn't trigger VMI re-creation. Expected: %s, Got: %s", *vmiUid, vmi.UID)
		}
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
	c, err := testAccProvider.Meta().(*config.Config).K8sClient()
	if err != nil {
		return nil, err
	}

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

			c, err := testAccProvider.Meta().(*config.Config).K8sClient()
			if err != nil {
				return err
			}
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

func enableCPUManager(t *testing.T, ctx context.Context, nodeName string) {
	if err := updateCPUManagerPolicy(ctx, nodeName, true); err != nil {
		t.Fatalf("failed to enable cpu manager: %v", err)
	}
}

func updateCPUManagerPolicy(ctx context.Context, nodeName string, enableCPUManager bool) error {
	c, err := testAccProvider.Meta().(*config.Config).K8sClient()
	if err != nil {
		return err
	}
	action := "disableCPUManager"
	if enableCPUManager {
		action = "enableCPUManager"
	}
	apiURL, err := buildNodeActionURL(c.RestConfig, nodeName, action)
	if err != nil {
		return fmt.Errorf("failed to build node action url: %v", err)
	}
	req, err := createRequest(apiURL, c.RestConfig.BearerToken)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // #nosec G402
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response, status code: %d, err: %v", resp.StatusCode, err)
		}
		return fmt.Errorf("status code: %d, err: %s", resp.StatusCode, string(bodyBytes))
	}

	err = waitForCPUMangerLabel(ctx, c, nodeName, enableCPUManager)
	if err != nil {
		return fmt.Errorf("wait cpu manager label failed: %v", err)
	}
	return nil
}

func buildNodeActionURL(config *rest.Config, nodeName, action string) (string, error) {
	parsedURL, err := url.Parse(config.Host)
	if err != nil {
		return "", fmt.Errorf("failed to parse restconfig host to url: %v", err)
	}

	parsedURL.Path = "/v1/harvester/nodes/" + nodeName
	query := parsedURL.Query()
	query.Set("action", action)
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), nil
}

func createRequest(apiURL, bearerToken string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	return req, nil
}

func waitForCPUMangerLabel(ctx context.Context, c *client.Client, nodeName string, enableCPUManager bool) error {
	return waitUntil(func() (bool, error) {
		node, err := c.KubeClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}

		expectedValue := strconv.FormatBool(enableCPUManager)
		if value, exists := node.Labels["cpumanager"]; exists && value == expectedValue {
			return true, nil
		}
		return false, nil
	})
}

func waitUntil(fun func() (bool, error)) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout reached, the condition was not met within 5 minutes")
		case <-ticker.C:
			ok, err := fun()
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
		}
	}
}
