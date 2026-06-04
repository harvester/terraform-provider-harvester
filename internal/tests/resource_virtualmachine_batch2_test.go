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

func TestAccVirtualMachine_access_credentials(t *testing.T) {
	var (
		testAccName         = "test-acc-acred-" + uuid.New().String()[:6]
		testAccNamespace    = "default"
		testAccResourceName = constants.ResourceTypeVirtualMachine + "." + testAccName
		secretName          = "test-ssh-keys-" + uuid.New().String()[:6]
		vm                  = &kubevirtv1.VirtualMachine{}
		ctx                 = context.Background()
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCreateSecret(t, ctx, testAccNamespace, secretName, map[string][]byte{
				"key1": []byte("ssh-ed25519 AAAA... test@example.com"),
			})
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

	access_credentials {
		ssh_public_key {
			secret_name        = "%s"
			propagation_method = "noCloud"
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
`, constants.ResourceTypeVirtualMachine, testAccName, testAccName, secretName, fedoraCloudContainer),
				Check: resource.ComposeTestCheckFunc(
					testAccVirtualMachineExists(ctx, testAccResourceName, vm),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineAccessCredentials+".#", "1"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineAccessCredentials+".0.ssh_public_key.#", "1"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineAccessCredentials+".0.ssh_public_key.0.secret_name", secretName),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineAccessCredentials+".0.ssh_public_key.0.propagation_method", "noCloud"),
					testAccCheckVMAccessCredentials(ctx, testAccResourceName),
				),
			},
		},
	})
}

func TestAccVirtualMachine_configmap_disk(t *testing.T) {
	var (
		testAccName         = "test-acc-cmdisk-" + uuid.New().String()[:6]
		testAccNamespace    = "default"
		testAccResourceName = constants.ResourceTypeVirtualMachine + "." + testAccName
		configMapName       = "test-cm-" + uuid.New().String()[:6]
		vm                  = &kubevirtv1.VirtualMachine{}
		ctx                 = context.Background()
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCreateConfigMap(t, ctx, testAccNamespace, configMapName, map[string]string{
				"config.yaml": "key: value",
			})
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
		name           = "config-disk"
		type           = "disk"
		bus            = "virtio"
		configmap_name = "%s"
	}
}
`, constants.ResourceTypeVirtualMachine, testAccName, testAccName, fedoraCloudContainer, configMapName),
				Check: resource.ComposeTestCheckFunc(
					testAccVirtualMachineExists(ctx, testAccResourceName, vm),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineDisk+".#", "2"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineDisk+".1.configmap_name", configMapName),
					testAccCheckVMConfigMapVolume(ctx, testAccResourceName, configMapName),
				),
			},
		},
	})
}

// testAccCreateSecret creates a K8s secret as a test prerequisite.
func testAccCreateSecret(t *testing.T, ctx context.Context, namespace, name string, data map[string][]byte) {
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
		Data: data,
	}
	_, err = c.KubeClient.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create test secret %s: %v", name, err)
	}
	t.Cleanup(func() {
		_ = c.KubeClient.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{}) //nolint:errcheck
	})
}

// testAccCreateConfigMap creates a K8s ConfigMap as a test prerequisite.
func testAccCreateConfigMap(t *testing.T, ctx context.Context, namespace, name string, data map[string]string) {
	t.Helper()
	c, err := testAccProvider.Meta().(*config.Config).K8sClient()
	if err != nil {
		t.Fatalf("failed to get k8s client: %v", err)
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
	_, err = c.KubeClient.CoreV1().ConfigMaps(namespace).Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create test configmap %s: %v", name, err)
	}
	t.Cleanup(func() {
		_ = c.KubeClient.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{}) //nolint:errcheck
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

// testAccCheckVMAccessCredentials verifies the K8s VM object has access credentials.
func testAccCheckVMAccessCredentials(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vm, err := testAccGetVirtualMachine(ctx, s, n)
		if err != nil {
			return err
		}
		if vm.Spec.Template == nil {
			return fmt.Errorf("VM template is nil")
		}
		creds := vm.Spec.Template.Spec.AccessCredentials
		if len(creds) != 1 {
			return fmt.Errorf("expected 1 access credential, got %d", len(creds))
		}
		if creds[0].SSHPublicKey == nil {
			return fmt.Errorf("expected SSH public key credential, got nil")
		}
		if creds[0].SSHPublicKey.Source.Secret == nil {
			return fmt.Errorf("expected SSH public key source secret, got nil")
		}
		if creds[0].SSHPublicKey.PropagationMethod.NoCloud == nil {
			return fmt.Errorf("expected noCloud propagation method")
		}
		return nil
	}
}

// testAccCheckVMConfigMapVolume verifies the K8s VM object has a ConfigMap volume.
func testAccCheckVMConfigMapVolume(ctx context.Context, n, configMapName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vm, err := testAccGetVirtualMachine(ctx, s, n)
		if err != nil {
			return err
		}
		if vm.Spec.Template == nil {
			return fmt.Errorf("VM template is nil")
		}
		for _, vol := range vm.Spec.Template.Spec.Volumes {
			if vol.ConfigMap != nil && vol.ConfigMap.Name == configMapName {
				return nil
			}
		}
		return fmt.Errorf("configmap volume %s not found in VM volumes", configMapName)
	}
}
