package tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/harvester/harvester/pkg/builder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
)

// testAccAffinityNodeName returns the hostname label of a node of the cluster,
// so the required node affinity of the test can be satisfied.
func testAccAffinityNodeName(ctx context.Context, t *testing.T) string {
	testAccPreCheck(t)
	c, err := testAccProvider.Meta().(*config.Config).K8sClient()
	if err != nil {
		t.Fatal(err)
	}
	nodes, err := c.KubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil || len(nodes.Items) == 0 {
		t.Fatalf("cannot list nodes: %v", err)
	}
	return nodes.Items[0].Labels[corev1.LabelHostname]
}

// testAccVirtualMachineAffinityConfig returns the test VM, with affinity
// blocks when weight is not zero.
func testAccVirtualMachineAffinityConfig(nodeName string, weight int) string {
	affinity := ""
	if weight != 0 {
		affinity = fmt.Sprintf(`
	node_affinity {
		required {
			node_selector_term {
				match_expressions {
					key      = %q
					operator = "In"
					values   = [%q]
				}
			}
		}
	}

	pod_affinity {
		preferred {
			weight = %d
			pod_affinity_term {
				topology_key = %q
				label_selector {
					match_labels = { app = "cache" }
				}
			}
		}
	}

	pod_anti_affinity {
		required {
			topology_key = %q
			label_selector {
				match_expressions {
					key      = "app"
					operator = "In"
					values   = ["web"]
				}
			}
		}
	}
`, corev1.LabelHostname, nodeName, weight, corev1.LabelHostname, corev1.LabelHostname)
	}
	return fmt.Sprintf(`
resource "harvester_virtualmachine" "test-acc-affinity" {
	name      = "test-acc-affinity"
	namespace = "default"

	cpu          = 1
	memory       = "1Gi"
	run_strategy = "Halted"
	machine_type = "q35"

	network_interface {
		name = "default"
	}

	disk {
		name       = "rootdisk"
		type       = "disk"
		bus        = "virtio"
		boot_order = 1

		container_image_name = %q
	}
%s
}
`, fedoraCloudContainer, affinity)
}

// testAccVirtualMachineUserAffinity checks whether the VM carries the user
// node affinity and required pod anti-affinity of the test.
func testAccVirtualMachineUserAffinity(ctx context.Context, n, nodeName string, expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vm, err := testAccGetVirtualMachine(ctx, s, n)
		if err != nil {
			return err
		}
		affinity := vm.Spec.Template.Spec.Affinity
		hasNode, hasAnti := false, false
		if affinity != nil && affinity.NodeAffinity != nil && affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			for _, term := range affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
				for _, expr := range term.MatchExpressions {
					hasNode = hasNode || (expr.Key == corev1.LabelHostname && len(expr.Values) == 1 && expr.Values[0] == nodeName)
				}
			}
		}
		if affinity != nil && affinity.PodAntiAffinity != nil {
			hasAnti = len(affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) == 1
		}
		if hasNode != expected || hasAnti != expected {
			return fmt.Errorf("user node affinity on the VM: %v, user pod anti-affinity: %v, want %v", hasNode, hasAnti, expected)
		}
		return nil
	}
}

// testAccVirtualMachineDefaultAntiAffinity checks that the VM carries the
// preferred anti-affinity term the provider always sets with
// builder.DefaultPodAntiAffinity().
func testAccVirtualMachineDefaultAntiAffinity(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vm, err := testAccGetVirtualMachine(ctx, s, n)
		if err != nil {
			return err
		}
		affinity := vm.Spec.Template.Spec.Affinity
		found := false
		if affinity != nil && affinity.PodAntiAffinity != nil {
			for _, term := range affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
				found = found || (term.PodAffinityTerm.LabelSelector != nil && isCreatorSelector(term.PodAffinityTerm.LabelSelector))
			}
		}
		if !found {
			return fmt.Errorf("default pod anti-affinity on %s is missing: %+v", builder.LabelKeyVirtualMachineCreator, affinity)
		}
		return nil
	}
}

func isCreatorSelector(selector *metav1.LabelSelector) bool {
	for _, expr := range selector.MatchExpressions {
		if expr.Key == builder.LabelKeyVirtualMachineCreator {
			return true
		}
	}
	return false
}

// TestAccVirtualMachine_affinity checks that node affinity, pod affinity and
// pod anti-affinity rules can be set, changed and removed, and that the default
// anti-affinity stays on the VM. Each step also fails if the plan after apply
// is not empty.
func TestAccVirtualMachine_affinity(t *testing.T) {
	const resourceName = "harvester_virtualmachine.test-acc-affinity"
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skip("Skipping test: TF_ACC is not set")
	}
	ctx := context.Background()
	// The configuration needs a real node name before the steps are built.
	nodeName := testAccAffinityNodeName(ctx, t)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineAffinityConfig(nodeName, 20),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "node_affinity.0.required.0.node_selector_term.0.match_expressions.0.values.0", nodeName),
					resource.TestCheckResourceAttr(resourceName, "pod_affinity.0.preferred.0.weight", "20"),
					resource.TestCheckResourceAttr(resourceName, "pod_anti_affinity.0.required.0.topology_key", corev1.LabelHostname),
					testAccVirtualMachineUserAffinity(ctx, resourceName, nodeName, true),
					testAccVirtualMachineDefaultAntiAffinity(ctx, resourceName),
				),
			},
			{
				Config: testAccVirtualMachineAffinityConfig(nodeName, 30),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "pod_affinity.0.preferred.0.weight", "30"),
					testAccVirtualMachineUserAffinity(ctx, resourceName, nodeName, true),
					testAccVirtualMachineDefaultAntiAffinity(ctx, resourceName),
				),
			},
			{
				Config: testAccVirtualMachineAffinityConfig(nodeName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "node_affinity.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "pod_anti_affinity.#", "0"),
					testAccVirtualMachineUserAffinity(ctx, resourceName, nodeName, false),
					testAccVirtualMachineDefaultAntiAffinity(ctx, resourceName),
				),
			},
		},
	})
}
