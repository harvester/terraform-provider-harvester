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

func TestAccVirtualMachine_secret_disk(t *testing.T) {
	var (
		testAccName         = "test-acc-secdisk-" + uuid.New().String()[:6]
		testAccNamespace    = "default"
		testAccResourceName = constants.ResourceTypeVirtualMachine + "." + testAccName
		secretName          = "test-secret-" + uuid.New().String()[:6]
		vm                  = &kubevirtv1.VirtualMachine{}
		ctx                 = context.Background()
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCreateSecret(t, ctx, testAccNamespace, secretName, map[string][]byte{
				"credentials": []byte("user=admin"),
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
		name        = "secret-disk"
		type        = "disk"
		bus         = "virtio"
		secret_name = "%s"
	}
}
`, constants.ResourceTypeVirtualMachine, testAccName, testAccName, fedoraCloudContainer, secretName),
				Check: resource.ComposeTestCheckFunc(
					testAccVirtualMachineExists(ctx, testAccResourceName, vm),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineDisk+".#", "2"),
					resource.TestCheckResourceAttr(testAccResourceName, constants.FieldVirtualMachineDisk+".1.secret_name", secretName),
					testAccCheckVMSecretVolume(ctx, testAccResourceName, secretName),
				),
			},
		},
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

// testAccCreateSecret creates a K8s Secret as a test prerequisite.
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

// testAccCheckVMConfigMapVolume verifies the K8s VM object mounts the ConfigMap as a volume.
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

// testAccCheckVMSecretVolume verifies the K8s VM object mounts the Secret as a volume.
func testAccCheckVMSecretVolume(ctx context.Context, n, secretName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vm, err := testAccGetVirtualMachine(ctx, s, n)
		if err != nil {
			return err
		}
		if vm.Spec.Template == nil {
			return fmt.Errorf("VM template is nil")
		}
		for _, vol := range vm.Spec.Template.Spec.Volumes {
			if vol.Secret != nil && vol.Secret.SecretName == secretName {
				return nil
			}
		}
		return fmt.Errorf("secret volume %s not found in VM volumes", secretName)
	}
}
