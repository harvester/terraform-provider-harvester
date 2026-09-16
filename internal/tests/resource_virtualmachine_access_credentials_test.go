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

// testAccCheckVMAccessCredentials verifies the K8s VM object carries the access credential.
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
