package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

func TestAccVirtualMachine_hyperv(t *testing.T) {
	var (
		testAccName         = "test-acc-hyperv-" + uuid.New().String()[:6]
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

	hyperv {
		relaxed           = true
		vapic             = true
		vpindex           = true
		runtime           = true
		synic             = true
		reset             = true
		frequencies       = true
		reenlightenment   = true
		tlbflush          = true
		ipi               = true
		evmcs             = false
		spinlocks         = true
		spinlocks_retries = 8192
		synictimer        = true
		synictimer_direct = true
		vendorid          = true
		vendorid_value    = "KVMKVMKVM"
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
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineHyperv+".#", "1"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineHyperv+".0.relaxed", "true"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineHyperv+".0.vapic", "true"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineHyperv+".0.vpindex", "true"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineHyperv+".0.spinlocks", "true"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineHyperv+".0.spinlocks_retries", "8192"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineHyperv+".0.synictimer", "true"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineHyperv+".0.synictimer_direct", "true"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineHyperv+".0.vendorid", "true"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineHyperv+".0.vendorid_value", "KVMKVMKVM"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineHypervPassthrough, "false"),
					testAccCheckVMHyperv(ctx, testAccResourceName),
				),
			},
		},
	})
}

func TestAccVirtualMachine_hyperv_passthrough(t *testing.T) {
	var (
		testAccName         = "test-acc-hvpt-" + uuid.New().String()[:6]
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

	hyperv_passthrough = true

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
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineHypervPassthrough, "true"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineHyperv+".#", "0"),
					testAccCheckVMHypervPassthrough(ctx, testAccResourceName),
				),
			},
		},
	})
}

func TestAccVirtualMachine_clock(t *testing.T) {
	var (
		testAccName         = "test-acc-clock-" + uuid.New().String()[:6]
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

	clock {
		timezone = "Europe/Paris"

		timer {
			hpet {
				enabled     = false
				tick_policy = "delay"
			}
			pit {
				enabled     = true
				tick_policy = "delay"
			}
			rtc {
				enabled     = true
				tick_policy = "catchup"
				track       = "guest"
			}
			hyperv {
				enabled = true
			}
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
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineClock+".#", "1"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineClock+".0.timezone", "Europe/Paris"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineClock+".0.timer.#", "1"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineClock+".0.timer.0.hpet.#", "1"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineClock+".0.timer.0.hpet.0.enabled", "false"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineClock+".0.timer.0.hpet.0.tick_policy", "delay"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineClock+".0.timer.0.pit.#", "1"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineClock+".0.timer.0.pit.0.tick_policy", "delay"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineClock+".0.timer.0.rtc.#", "1"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineClock+".0.timer.0.rtc.0.tick_policy", "catchup"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineClock+".0.timer.0.rtc.0.track", "guest"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineClock+".0.timer.0.hyperv.#", "1"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineClock+".0.timer.0.hyperv.0.enabled", "true"),
					testAccCheckVMClock(ctx, testAccResourceName),
				),
			},
		},
	})
}

func TestAccVirtualMachine_sysprep_secret(t *testing.T) {
	var (
		testAccName         = "test-acc-sysprep-" + uuid.New().String()[:6]
		testAccNamespace    = "default"
		testAccResourceName = constants.ResourceTypeVirtualMachine + "." + testAccName
		sysprepSecretName   = "test-sysprep-" + uuid.New().String()[:6]
		vm                  = &kubevirtv1.VirtualMachine{}
		ctx                 = context.Background()
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCreateSysprepSecret(t, ctx, testAccNamespace, sysprepSecretName)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource %s "%s" {
	name = "%s"

	cpu    = 1
	memory = "1Gi"

	run_strategy = "RerunOnFailure"
	machine_type = "q35"

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

	disk {
		name                = "sysprep-disk"
		type                = "disk"
		bus                 = "virtio"
		sysprep_secret_name = "%s"
	}
}
`, constants.ResourceTypeVirtualMachine, testAccName, testAccName, fedoraCloudContainer, sysprepSecretName),
				Check: resource.ComposeTestCheckFunc(
					testAccVirtualMachineExists(ctx, testAccResourceName, vm),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineDisk+".#", "2"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineDisk+".1.sysprep_secret_name", sysprepSecretName),
					testAccCheckVMSysprepVolume(ctx, testAccResourceName, sysprepSecretName, "secret"),
				),
			},
		},
	})
}

// testAccCreateSysprepSecret creates a K8s secret with sysprep content.
func testAccCreateSysprepSecret(t *testing.T, ctx context.Context, namespace, name string) {
	t.Helper()
	c, err := testAccProvider.Meta().(*config.Config).K8sClient()
	if err != nil {
		t.Fatalf("failed to get k8s client: %v", err)
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"autounattend.xml": []byte("<unattend><settings></settings></unattend>"),
		},
	}
	_, err = c.KubeClient.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create test sysprep secret %s: %v", name, err)
	}
	t.Cleanup(func() {
		_ = c.KubeClient.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{}) //nolint:errcheck
	})
}

// testAccCheckVMHyperv verifies the K8s VM object has Hyper-V features set.
func testAccCheckVMHyperv(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vm, err := testAccGetVirtualMachine(ctx, s, n)
		if err != nil {
			return err
		}
		if vm.Spec.Template == nil || vm.Spec.Template.Spec.Domain.Features == nil {
			return fmt.Errorf("VM features are nil")
		}
		hv := vm.Spec.Template.Spec.Domain.Features.Hyperv
		if hv == nil {
			return fmt.Errorf("Hyper-V features are nil")
		}
		if hv.Relaxed == nil || hv.Relaxed.Enabled == nil || !*hv.Relaxed.Enabled {
			return fmt.Errorf("expected Relaxed to be enabled")
		}
		if hv.VAPIC == nil || hv.VAPIC.Enabled == nil || !*hv.VAPIC.Enabled {
			return fmt.Errorf("expected VAPIC to be enabled")
		}
		if hv.Spinlocks == nil || hv.Spinlocks.Enabled == nil || !*hv.Spinlocks.Enabled {
			return fmt.Errorf("expected Spinlocks to be enabled")
		}
		if hv.Spinlocks.Retries == nil || *hv.Spinlocks.Retries != 8192 {
			return fmt.Errorf("expected Spinlocks retries to be 8192")
		}
		if hv.SyNICTimer == nil || hv.SyNICTimer.Enabled == nil || !*hv.SyNICTimer.Enabled {
			return fmt.Errorf("expected SyNICTimer to be enabled")
		}
		if hv.SyNICTimer.Direct == nil || hv.SyNICTimer.Direct.Enabled == nil || !*hv.SyNICTimer.Direct.Enabled {
			return fmt.Errorf("expected SyNICTimer Direct to be enabled")
		}
		if hv.VendorID == nil || hv.VendorID.Enabled == nil || !*hv.VendorID.Enabled {
			return fmt.Errorf("expected VendorID to be enabled")
		}
		if hv.VendorID.VendorID != "KVMKVMKVM" {
			return fmt.Errorf("expected VendorID value KVMKVMKVM, got %s", hv.VendorID.VendorID)
		}
		return nil
	}
}

// testAccCheckVMHypervPassthrough verifies HypervPassthrough is enabled.
func testAccCheckVMHypervPassthrough(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vm, err := testAccGetVirtualMachine(ctx, s, n)
		if err != nil {
			return err
		}
		if vm.Spec.Template == nil || vm.Spec.Template.Spec.Domain.Features == nil {
			return fmt.Errorf("VM features are nil")
		}
		hp := vm.Spec.Template.Spec.Domain.Features.HypervPassthrough
		if hp == nil || hp.Enabled == nil || !*hp.Enabled {
			return fmt.Errorf("expected HypervPassthrough to be enabled")
		}
		if vm.Spec.Template.Spec.Domain.Features.Hyperv != nil {
			return fmt.Errorf("expected Hyperv to be nil when using passthrough")
		}
		return nil
	}
}

// testAccCheckVMClock verifies the K8s VM object has the expected clock configuration.
func testAccCheckVMClock(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vm, err := testAccGetVirtualMachine(ctx, s, n)
		if err != nil {
			return err
		}
		if vm.Spec.Template == nil {
			return fmt.Errorf("VM template is nil")
		}
		clock := vm.Spec.Template.Spec.Domain.Clock
		if clock == nil {
			return fmt.Errorf("clock is nil")
		}
		if clock.Timezone == nil {
			return fmt.Errorf("timezone is nil")
		}
		tz := string(*clock.Timezone)
		if tz != "Europe/Paris" {
			return fmt.Errorf("expected timezone Europe/Paris, got %s", tz)
		}
		if clock.Timer == nil {
			return fmt.Errorf("timer is nil")
		}
		if clock.Timer.HPET == nil {
			return fmt.Errorf("HPET timer is nil")
		}
		if clock.Timer.HPET.Enabled == nil || *clock.Timer.HPET.Enabled {
			return fmt.Errorf("expected HPET to be disabled")
		}
		if clock.Timer.RTC == nil {
			return fmt.Errorf("RTC timer is nil")
		}
		if clock.Timer.RTC.Track != kubevirtv1.TrackGuest {
			return fmt.Errorf("expected RTC track guest, got %s", clock.Timer.RTC.Track)
		}
		if clock.Timer.Hyperv == nil {
			return fmt.Errorf("Hyperv timer is nil")
		}
		if clock.Timer.Hyperv.Enabled == nil || !*clock.Timer.Hyperv.Enabled {
			return fmt.Errorf("expected Hyperv timer to be enabled")
		}
		return nil
	}
}

// testAccCheckVMSysprepVolume verifies the K8s VM object has a sysprep volume.
func testAccCheckVMSysprepVolume(ctx context.Context, n, sysprepName, sourceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vm, err := testAccGetVirtualMachine(ctx, s, n)
		if err != nil {
			return err
		}
		if vm.Spec.Template == nil {
			return fmt.Errorf("VM template is nil")
		}
		for _, vol := range vm.Spec.Template.Spec.Volumes {
			if vol.Sysprep != nil {
				if sourceType == "secret" && vol.Sysprep.Secret != nil && vol.Sysprep.Secret.Name == sysprepName {
					return nil
				}
				if sourceType == "configmap" && vol.Sysprep.ConfigMap != nil && vol.Sysprep.ConfigMap.Name == sysprepName {
					return nil
				}
			}
		}
		return fmt.Errorf("sysprep volume with %s %s not found", sourceType, sysprepName)
	}
}
