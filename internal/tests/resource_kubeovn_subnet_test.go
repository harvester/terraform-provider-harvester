package tests

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"testing"

	"github.com/harvester/harvester-network-controller/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	nadv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	kubeovnclient "github.com/kubeovn/kube-ovn/pkg/client/clientset/versioned"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

const (
	testAccKubeOVNVpcName         = "test-acc-ko-vpc"
	testAccKubeOVNSubnetName      = "test-acc-ko-subnet"
	testAccKubeOVNSubnetIPv6Name  = "test-acc-ko-subnet6"
	testAccKubeOVNNetworkName     = "test-acc-ko-net"
	testAccKubeOVNNetworkIPv6Name = "test-acc-ko-net6"
	testAccKubeOVNGateway         = "10.242.0.1"
	testAccKubeOVNExcludedIP      = "10.242.0.50"
	testAccKubeOVNDHCPOptions     = "lease_time=3600"
)

func testAccKubeOVNClient(t *testing.T) *kubeovnclient.Clientset {
	c, err := testAccProvider.Meta().(*config.Config).K8sClient()
	if err != nil {
		t.Fatal(err)
	}
	return c.KubeOVNClient
}

// testAccKubeOVNSubnetPreCheck creates the overlay networks the subnets are
// attached to: the Harvester network webhook refuses a kube-ovn subnet
// without a provider, and a provider serves a single subnet. The test is
// skipped when KubeOVN is not installed.
func testAccKubeOVNSubnetPreCheck(ctx context.Context, t *testing.T) {
	testAccPreCheck(t)
	if _, err := testAccKubeOVNClient(t).KubeovnV1().Subnets().List(ctx, metav1.ListOptions{Limit: 1}); err != nil {
		t.Skipf("KubeOVN is not available: %v", err)
	}
	c, err := testAccProvider.Meta().(*config.Config).K8sClient()
	if err != nil {
		t.Fatal(err)
	}
	nads := c.HarvesterClient.K8sCniCncfIoV1().NetworkAttachmentDefinitions("default")
	for _, name := range []string{testAccKubeOVNNetworkName, testAccKubeOVNNetworkIPv6Name} {
		nad := &nadv1.NetworkAttachmentDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "default",
				Labels:    map[string]string{utils.KeyNetworkType: string(utils.OverlayNetwork)},
			},
			Spec: nadv1.NetworkAttachmentDefinitionSpec{
				Config: fmt.Sprintf(`{"cniVersion":"0.3.1","type":%q,"server_socket":"/run/openvswitch/kube-ovn-daemon.sock","provider":"%s.default.ovn"}`, utils.CNITypeKubeOVN, name),
			},
		}
		if _, err := nads.Create(ctx, nad, metav1.CreateOptions{}); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := nads.Delete(context.Background(), name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
				t.Error(err)
			}
		})
	}
}

// testAccKubeOVNSubnetConfig declares the VPC, an IPv4 subnet whose options
// vary with the step, and an IPv6 subnet without protocol, which kube-ovn
// derives from the CIDR.
func testAccKubeOVNSubnetConfig(enabled bool, excludeIPs, options string) string {
	return fmt.Sprintf(`
resource "harvester_kubeovn_vpc" "test" {
	name       = %q
	namespaces = ["default"]
}

resource "harvester_kubeovn_subnet" "test" {
	name        = %q
	vpc         = harvester_kubeovn_vpc.test.name
	cidr_block  = "10.242.0.0/24"
	gateway     = %q
	exclude_ips = [%s]

	enable_dhcp  = %t
	nat_outgoing = %t
	enable_lb    = %t
	%s
}

resource "harvester_kubeovn_subnet" "test6" {
	name             = %q
	vpc              = harvester_kubeovn_vpc.test.name
	cidr_block       = "fd00:10:242::/64"
	gateway          = "fd00:10:242::1"
	network_provider = "%s.default.ovn"
}
`, testAccKubeOVNVpcName, testAccKubeOVNSubnetName, testAccKubeOVNGateway, excludeIPs, enabled, enabled, enabled, options,
		testAccKubeOVNSubnetIPv6Name, testAccKubeOVNNetworkIPv6Name)
}

func testAccKubeOVNSubnetProvider() string {
	return fmt.Sprintf("network_provider = \"%s.default.ovn\"", testAccKubeOVNNetworkName)
}

// testAccKubeOVNSubnetSpec checks the booleans, the DHCP options and the
// excluded IPs stored by kube-ovn, the gateway being always excluded.
func testAccKubeOVNSubnetSpec(ctx context.Context, t *testing.T, enabled bool, dhcpOptions string, excludeIPs ...string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		subnet, err := testAccKubeOVNClient(t).KubeovnV1().Subnets().Get(ctx, testAccKubeOVNSubnetName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		spec := subnet.Spec
		if spec.EnableDHCP != enabled || spec.NatOutgoing != enabled || spec.EnableLb == nil || *spec.EnableLb != enabled {
			return fmt.Errorf("enableDHCP = %v, natOutgoing = %v, enableLb = %v, want %v", spec.EnableDHCP, spec.NatOutgoing, spec.EnableLb, enabled)
		}
		if spec.DHCPv4Options != dhcpOptions {
			return fmt.Errorf("dhcpV4Options = %q, want %q", spec.DHCPv4Options, dhcpOptions)
		}
		for _, ip := range append(excludeIPs, testAccKubeOVNGateway) {
			if !slices.Contains(spec.ExcludeIps, ip) {
				return fmt.Errorf("excludeIps = %v, want %s in it", spec.ExcludeIps, ip)
			}
		}
		return nil
	}
}

// testAccKubeOVNVpcReady checks that the VPC was ready when the apply
// returned: kube-ovn reports it with status.standby.
func testAccKubeOVNVpcReady(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		vpc, err := testAccKubeOVNClient(t).KubeovnV1().Vpcs().Get(ctx, testAccKubeOVNVpcName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if !vpc.Status.Standby {
			return fmt.Errorf("vpc %s is not ready (status.standby = false)", vpc.Name)
		}
		return nil
	}
}

// testAccCheckKubeOVNSubnetDestroy checks, without waiting, that the subnets
// and the VPC are gone: the destroy must only return once their kube-ovn
// finalizers are released.
func testAccCheckKubeOVNSubnetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := testAccKubeOVNClient(t).KubeovnV1()
		for _, name := range []string{testAccKubeOVNSubnetName, testAccKubeOVNSubnetIPv6Name} {
			if _, err := client.Subnets().Get(ctx, name, metav1.GetOptions{}); !apierrors.IsNotFound(err) {
				return fmt.Errorf("subnet %s still exists after destroy (err = %v)", name, err)
			}
		}
		if _, err := client.Vpcs().Get(ctx, testAccKubeOVNVpcName, metav1.GetOptions{}); !apierrors.IsNotFound(err) {
			return fmt.Errorf("vpc %s still exists after destroy (err = %v)", testAccKubeOVNVpcName, err)
		}
		return nil
	}
}

// TestAccKubeOVNSubnet_basic creates a subnet with every boolean on and DHCP
// options, then switches the booleans off, removes the DHCP options and adds
// an excluded IP, then removes network_provider, which the Harvester webhook
// must refuse instead of the change being silently skipped. Each step also
// fails if the plan after apply is not empty, which covers the gateway the
// provider adds to the excluded IPs and the protocol kube-ovn derives for the
// IPv6 subnet.
func TestAccKubeOVNSubnet_basic(t *testing.T) {
	const (
		resourceName     = constants.ResourceTypeKubeOVNSubnet + ".test"
		ipv6ResourceName = constants.ResourceTypeKubeOVNSubnet + ".test6"
	)
	ctx := context.Background()
	dhcpOptions := fmt.Sprintf("dhcp_v4_options = %q", testAccKubeOVNDHCPOptions)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccKubeOVNSubnetPreCheck(ctx, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubeOVNSubnetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKubeOVNSubnetConfig(true, "", testAccKubeOVNSubnetProvider()+"\n"+dhcpOptions),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, constants.FieldCommonState, constants.StateCommonReady),
					resource.TestCheckResourceAttr(ipv6ResourceName, constants.FieldCommonState, constants.StateCommonReady),
					resource.TestCheckResourceAttr(ipv6ResourceName, constants.FieldKubeOVNSubnetProtocol, kubeovnv1.ProtocolIPv6),
					resource.TestCheckResourceAttr(resourceName, constants.FieldKubeOVNSubnetEnableDHCP, "true"),
					resource.TestCheckTypeSetElemAttr(resourceName, constants.FieldKubeOVNSubnetExcludeIPs+".*", testAccKubeOVNGateway),
					testAccKubeOVNVpcReady(ctx, t),
					testAccKubeOVNSubnetSpec(ctx, t, true, testAccKubeOVNDHCPOptions),
				),
			},
			{
				Config: testAccKubeOVNSubnetConfig(false, fmt.Sprintf("%q", testAccKubeOVNExcludedIP), testAccKubeOVNSubnetProvider()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, constants.FieldKubeOVNSubnetEnableDHCP, "false"),
					resource.TestCheckResourceAttr(resourceName, constants.FieldKubeOVNSubnetEnableLb, "false"),
					resource.TestCheckResourceAttr(resourceName, constants.FieldKubeOVNSubnetDHCPv4Options, ""),
					testAccKubeOVNSubnetSpec(ctx, t, false, "", testAccKubeOVNExcludedIP),
				),
			},
			{
				Config:      testAccKubeOVNSubnetConfig(false, fmt.Sprintf("%q", testAccKubeOVNExcludedIP), ""),
				ExpectError: regexp.MustCompile("provider is empty"),
			},
		},
	})
}
