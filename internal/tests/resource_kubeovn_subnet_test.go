package tests

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"
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
		if _, err := nads.Create(ctx, nad, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := nads.Delete(context.Background(), name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
				t.Error(err)
			}
		})
	}
}

// testAccKubeOVNSubnetConfig declares the VPC with a static and a policy
// route, an IPv4 subnet whose options vary with the step, an IPv6 subnet
// without gateway nor protocol, which kube-ovn derives from the CIDR, and
// the data sources of the VPC and of the IPv4 subnet. excludeIPs is the
// whole exclude_ips line, so that a step can leave the attribute out.
func testAccKubeOVNSubnetConfig(enabled bool, excludeIPs, options string) string {
	return fmt.Sprintf(`
resource "harvester_kubeovn_vpc" "test" {
	name       = %q
	namespaces = ["default"]

	static_routes {
		cidr        = "192.0.2.0/24"
		next_hop_ip = "10.242.0.254"
	}

	policy_routes {
		priority = 10
		match    = "ip4.src == 10.242.0.0/24"
		action   = "allow"
	}
}

resource "harvester_kubeovn_subnet" "test" {
	name       = %q
	vpc        = harvester_kubeovn_vpc.test.name
	cidr_block = "10.242.0.0/24"
	gateway    = %q
	%s

	enable_dhcp  = %t
	nat_outgoing = %t
	enable_lb    = %t
	%s
}

resource "harvester_kubeovn_subnet" "test6" {
	name             = %q
	vpc              = harvester_kubeovn_vpc.test.name
	cidr_block       = "fd00:10:242::/64"
	network_provider = "%s.default.ovn"
}

data "harvester_kubeovn_vpc" "test" {
	name = harvester_kubeovn_vpc.test.name
}

data "harvester_kubeovn_subnet" "test" {
	name = harvester_kubeovn_subnet.test.name
}
`, testAccKubeOVNVpcName, testAccKubeOVNSubnetName, testAccKubeOVNGateway, excludeIPs, enabled, enabled, enabled, options,
		testAccKubeOVNSubnetIPv6Name, testAccKubeOVNNetworkIPv6Name)
}

func testAccKubeOVNExcludeIPs(entries ...string) string {
	quoted := make([]string, 0, len(entries))
	for _, entry := range entries {
		quoted = append(quoted, fmt.Sprintf("%q", entry))
	}
	return fmt.Sprintf("exclude_ips = [%s]", strings.Join(quoted, ", "))
}

func testAccKubeOVNSubnetProvider() string {
	return fmt.Sprintf("network_provider = \"%s.default.ovn\"", testAccKubeOVNNetworkName)
}

// testAccKubeOVNSubnetSpec checks the booleans, the DHCP options and the
// excluded IPs stored by kube-ovn, which must be exactly excludeIPs.
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
		if !slices.Equal(spec.ExcludeIps, excludeIPs) {
			return fmt.Errorf("excludeIps = %v, want %v", spec.ExcludeIps, excludeIPs)
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
// overlapping excluded ranges, then removes network_provider, which must be
// refused, then removes exclude_ips, then imports the VPC and the subnets.
// Each step also fails if the plan after apply is not empty, which covers
// the gateway kube-ovn adds to the excluded IPs, the ranges it merges, and
// the gateway and protocol it derives for the IPv6 subnet.
func TestAccKubeOVNSubnet_basic(t *testing.T) {
	const (
		vpcResourceName  = constants.ResourceTypeKubeOVNVpc + ".test"
		resourceName     = constants.ResourceTypeKubeOVNSubnet + ".test"
		ipv6ResourceName = constants.ResourceTypeKubeOVNSubnet + ".test6"
		excludedRange    = "10.242.0.20..10.242.0.30"
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
					resource.TestCheckResourceAttr(ipv6ResourceName, constants.FieldKubeOVNSubnetGateway, "fd00:10:242::1"),
					resource.TestCheckResourceAttr(resourceName, constants.FieldKubeOVNSubnetEnableDHCP, "true"),
					resource.TestCheckTypeSetElemAttr(resourceName, constants.FieldKubeOVNSubnetExcludeIPs+".*", testAccKubeOVNGateway),
					resource.TestCheckResourceAttr(vpcResourceName, constants.FieldCommonState, constants.StateCommonReady),
					resource.TestCheckResourceAttr(vpcResourceName, constants.FieldKubeOVNVpcStaticRoutes+".0."+constants.FieldKubeOVNStaticRoutePolicy, string(kubeovnv1.PolicyDst)),
					resource.TestCheckResourceAttr(vpcResourceName, constants.FieldKubeOVNVpcPolicyRoutes+".0."+constants.FieldKubeOVNPolicyRouteAction, string(kubeovnv1.PolicyRouteActionAllow)),
					resource.TestCheckResourceAttrPair("data."+vpcResourceName, constants.FieldKubeOVNVpcRouter, vpcResourceName, constants.FieldKubeOVNVpcRouter),
					resource.TestCheckResourceAttrPair("data."+resourceName, constants.FieldKubeOVNSubnetCIDRBlock, resourceName, constants.FieldKubeOVNSubnetCIDRBlock),
					testAccKubeOVNVpcReady(ctx, t),
					testAccKubeOVNSubnetSpec(ctx, t, true, testAccKubeOVNDHCPOptions, testAccKubeOVNGateway),
				),
			},
			{
				Config: testAccKubeOVNSubnetConfig(false, testAccKubeOVNExcludeIPs(testAccKubeOVNExcludedIP, "10.242.0.20..10.242.0.25", "10.242.0.24..10.242.0.30"), testAccKubeOVNSubnetProvider()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, constants.FieldKubeOVNSubnetEnableDHCP, "false"),
					resource.TestCheckResourceAttr(resourceName, constants.FieldKubeOVNSubnetEnableLb, "false"),
					resource.TestCheckResourceAttr(resourceName, constants.FieldKubeOVNSubnetDHCPv4Options, ""),
					testAccKubeOVNSubnetSpec(ctx, t, false, "", testAccKubeOVNGateway, excludedRange, testAccKubeOVNExcludedIP),
				),
			},
			{
				Config:      testAccKubeOVNSubnetConfig(false, "", ""),
				ExpectError: regexp.MustCompile(`"network_provider" is required`),
			},
			{
				Config: testAccKubeOVNSubnetConfig(false, "", testAccKubeOVNSubnetProvider()),
				Check:  testAccKubeOVNSubnetSpec(ctx, t, false, "", testAccKubeOVNGateway),
			},
			{
				ResourceName:      vpcResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      ipv6ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
