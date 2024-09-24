package tests

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
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
  %s = %s

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

func buildVirtualMachineConfig(name, description, memory string, cpuPinning bool) string {
	return fmt.Sprintf(testAccVirtualMachineConfigTemplate, constants.ResourceTypeVirtualMachine, name,
		constants.FieldCommonName, name,
		constants.FieldCommonDescription, description,
		constants.FieldVirtualMachineMemory, memory,
		constants.FieldVirtualMachineCPUPinning, strconv.FormatBool(cpuPinning))
}

func TestAccVirtualMachine_basic(t *testing.T) {
	var (
		vm  = &kubevirtv1.VirtualMachine{}
		ctx = context.Background()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: buildVirtualMachineConfig(testAccVirtualMachineName, testAccVirtualMachineDescription, testAccVirtualMachineMemory, false),
				Check: resource.ComposeTestCheckFunc(
					testAccVirtualMachineExists(ctx, testAccVirtualMachineResourceName, vm),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldCommonName, testAccVirtualMachineName),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldCommonDescription, testAccVirtualMachineDescription),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineMemory, testAccVirtualMachineMemory),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineCPUPinning, "false"),
				),
			},
			{
				Config: buildVirtualMachineConfig(testAccVirtualMachineName, testAccVirtualMachineDescription, testAccVirtualMachineMemoryUpdate, false),
				Check: resource.ComposeTestCheckFunc(
					testAccVirtualMachineExists(ctx, testAccVirtualMachineResourceName, vm),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldCommonName, testAccVirtualMachineName),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldCommonDescription, testAccVirtualMachineDescription),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineMemory, testAccVirtualMachineMemoryUpdate),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineCPUPinning, "false"),
				),
			},
		},
	})
}

func TestAccVirtualMachine_cpu_pinning(t *testing.T) {
	var (
		vm       = &kubevirtv1.VirtualMachine{}
		ctx      = context.Background()
		nodeName = ""
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			nodes, err := testAccProvider.Meta().(*client.Client).KubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if len(nodes.Items) == 0 {
				t.Fatal("failed to find any node")
			}
			nodeName = nodes.Items[0].Name
			t.Log("enable cpu manager on node " + nodeName)
			enableCPUManager(t, ctx, nodeName)
		},
		Providers: testAccProviders,
		CheckDestroy: func(s *terraform.State) error {
			err := testAccCheckVirtualMachineDestroy(ctx)(s)
			defer func() {
				t.Log("disable cpu manager on node " + nodeName)
				disableCPUManager(t, ctx, nodeName)
			}()
			return err
		},
		Steps: []resource.TestStep{
			{
				Config: buildVirtualMachineConfig(testAccVirtualMachineName, testAccVirtualMachineDescription, testAccVirtualMachineMemory, true),
				Check: resource.ComposeTestCheckFunc(
					testAccVirtualMachineExists(ctx, testAccVirtualMachineResourceName, vm),
					func(s *terraform.State) error {
						if vm.Spec.Template == nil || vm.Spec.Template.Spec.Domain.CPU == nil || !vm.Spec.Template.Spec.Domain.CPU.DedicatedCPUPlacement {
							return errors.New("DedicatedCPUPlacement should be true")
						}
						return nil
					},
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldCommonName, testAccVirtualMachineName),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldCommonDescription, testAccVirtualMachineDescription),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineMemory, testAccVirtualMachineMemory),
					resource.TestCheckResourceAttr(testAccVirtualMachineResourceName, constants.FieldVirtualMachineCPUPinning, "true"),
				),
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
		*vm = *foundVM
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

func enableCPUManager(t *testing.T, ctx context.Context, nodeName string) {
	if err := updateCPUManagerPolicy(ctx, nodeName, true); err != nil {
		t.Fatalf("failed to enable cpu manager: %v", err)
	}
}

func disableCPUManager(t *testing.T, ctx context.Context, nodeName string) {
	if err := updateCPUManagerPolicy(ctx, nodeName, false); err != nil {
		t.Fatalf("failed to disable cpu manager: %v", err)
	}
}

func updateCPUManagerPolicy(ctx context.Context, nodeName string, enableCPUManager bool) error {
	c := testAccProvider.Meta().(*client.Client)
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
	defer resp.Body.Close()

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
			return false, fmt.Errorf("failed to get node: %v", err)
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
