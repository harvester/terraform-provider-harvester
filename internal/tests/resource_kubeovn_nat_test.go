package tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/harvester/harvester-network-controller/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	nadv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

const (
	// Name of an existing kube-ovn subnet backed by a network attachment
	// definition, used as the external subnet of the NAT gateway, and a free
	// address of it for the EIP. The test is skipped when they are not set.
	testAccKubeOVNExternalSubnetEnv = "HARVESTER_KUBEOVN_EXTERNAL_SUBNET"
	testAccKubeOVNEIPIPEnv          = "HARVESTER_KUBEOVN_EIP_IP"

	testAccKubeOVNNatVpcName          = "test-acc-ko-nat-vpc"
	testAccKubeOVNNatSubnetName       = "test-acc-ko-nat-subnet"
	testAccKubeOVNNatNetworkName      = "test-acc-ko-nat-net"
	testAccKubeOVNNatNetworkNamespace = "default"
	testAccKubeOVNNatGatewayName      = "test-acc-ko-nat-gw"
	testAccKubeOVNNatEIPName          = "test-acc-ko-nat-eip"
	testAccKubeOVNNatSnatName         = "test-acc-ko-nat-snat"
	testAccKubeOVNNatDnatName         = "test-acc-ko-nat-dnat"
	testAccKubeOVNNatCIDR             = "10.243.0.0/24"
	testAccKubeOVNNatLanIP            = "10.243.0.254"
	testAccKubeOVNNatInternalIP       = "10.243.0.100"
)

// testAccKubeOVNNatPreCheck skips the test when KubeOVN or the external
// subnet is not available, and creates the overlay network the VPC subnet is
// attached to.
func testAccKubeOVNNatPreCheck(ctx context.Context, t *testing.T, externalSubnet, eipIP string) {
	if externalSubnet == "" || eipIP == "" {
		t.Skipf("%s and %s must be set to run the KubeOVN NAT acceptance test", testAccKubeOVNExternalSubnetEnv, testAccKubeOVNEIPIPEnv)
	}
	testAccPreCheck(t)
	if _, err := testAccKubeOVNClient(t).KubeovnV1().Subnets().Get(ctx, externalSubnet, metav1.GetOptions{}); err != nil {
		t.Skipf("external subnet %s is not available: %v", externalSubnet, err)
	}
	c, err := testAccProvider.Meta().(*config.Config).K8sClient()
	if err != nil {
		t.Fatal(err)
	}
	nads := c.HarvesterClient.K8sCniCncfIoV1().NetworkAttachmentDefinitions(testAccKubeOVNNatNetworkNamespace)
	nad := &nadv1.NetworkAttachmentDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testAccKubeOVNNatNetworkName,
			Namespace: testAccKubeOVNNatNetworkNamespace,
			Labels:    map[string]string{utils.KeyNetworkType: string(utils.OverlayNetwork)},
		},
		Spec: nadv1.NetworkAttachmentDefinitionSpec{
			Config: fmt.Sprintf(`{"cniVersion":"0.3.1","type":%q,"server_socket":"/run/openvswitch/kube-ovn-daemon.sock","provider":"%s.%s.ovn"}`, utils.CNITypeKubeOVN, testAccKubeOVNNatNetworkName, testAccKubeOVNNatNetworkNamespace),
		},
	}
	if _, err := nads.Create(ctx, nad, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := nads.Delete(context.Background(), testAccKubeOVNNatNetworkName, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			t.Error(err)
		}
	})
}

// testAccKubeOVNNatConfig declares a VPC that routes its external traffic to
// the NAT gateway, its subnet, the gateway, an EIP, an SNAT rule for the
// subnet and a DNAT rule left on the default protocol.
func testAccKubeOVNNatConfig(externalSubnet, eipIP string) string {
	return fmt.Sprintf(`
resource "harvester_kubeovn_vpc" "test" {
	name = %[1]q

	static_routes {
		cidr        = "0.0.0.0/0"
		next_hop_ip = %[5]q
	}
}

resource "harvester_kubeovn_subnet" "test" {
	name             = %[2]q
	vpc              = harvester_kubeovn_vpc.test.name
	cidr_block       = %[4]q
	network_provider = "%[3]s.default.ovn"
}

resource "harvester_kubeovn_vpc_nat_gateway" "test" {
	name             = %[6]q
	vpc              = harvester_kubeovn_vpc.test.name
	subnet           = harvester_kubeovn_subnet.test.name
	lan_ip           = %[5]q
	external_subnets = [%[7]q]
}

resource "harvester_kubeovn_iptables_eip" "test" {
	name            = %[8]q
	nat_gw_dp       = harvester_kubeovn_vpc_nat_gateway.test.name
	external_subnet = %[7]q
	v4_ip           = %[9]q
}

resource "harvester_kubeovn_iptables_snat_rule" "test" {
	name          = %[10]q
	eip           = harvester_kubeovn_iptables_eip.test.name
	internal_cidr = harvester_kubeovn_subnet.test.cidr_block
}

resource "harvester_kubeovn_iptables_dnat_rule" "test" {
	name          = %[11]q
	eip           = harvester_kubeovn_iptables_eip.test.name
	external_port = "8022"
	internal_ip   = %[12]q
	internal_port = "22"
}

data "harvester_kubeovn_iptables_eip" "test" {
	name = harvester_kubeovn_iptables_eip.test.name
}
`, testAccKubeOVNNatVpcName, testAccKubeOVNNatSubnetName, testAccKubeOVNNatNetworkName, testAccKubeOVNNatCIDR,
		testAccKubeOVNNatLanIP, testAccKubeOVNNatGatewayName, externalSubnet, testAccKubeOVNNatEIPName, eipIP,
		testAccKubeOVNNatSnatName, testAccKubeOVNNatDnatName, testAccKubeOVNNatInternalIP)
}

// testAccCheckKubeOVNNatDestroy checks, without waiting, that every object
// is gone: the destroy must only return once the kube-ovn finalizers of the
// rules and of the EIP are released, or the gateway, the subnet and the VPC
// could not have been deleted.
func testAccCheckKubeOVNNatDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := testAccKubeOVNClient(t).KubeovnV1()
		gets := map[string]func() error{
			testAccKubeOVNNatDnatName: func() error {
				_, err := client.IptablesDnatRules().Get(ctx, testAccKubeOVNNatDnatName, metav1.GetOptions{})
				return err
			},
			testAccKubeOVNNatSnatName: func() error {
				_, err := client.IptablesSnatRules().Get(ctx, testAccKubeOVNNatSnatName, metav1.GetOptions{})
				return err
			},
			testAccKubeOVNNatEIPName: func() error {
				_, err := client.IptablesEIPs().Get(ctx, testAccKubeOVNNatEIPName, metav1.GetOptions{})
				return err
			},
			testAccKubeOVNNatGatewayName: func() error {
				_, err := client.VpcNatGateways().Get(ctx, testAccKubeOVNNatGatewayName, metav1.GetOptions{})
				return err
			},
			testAccKubeOVNNatSubnetName: func() error {
				_, err := client.Subnets().Get(ctx, testAccKubeOVNNatSubnetName, metav1.GetOptions{})
				return err
			},
			testAccKubeOVNNatVpcName: func() error {
				_, err := client.Vpcs().Get(ctx, testAccKubeOVNNatVpcName, metav1.GetOptions{})
				return err
			},
		}
		for name, get := range gets {
			if err := get(); !apierrors.IsNotFound(err) {
				return fmt.Errorf("%s still exists after destroy (err = %v)", name, err)
			}
		}
		return nil
	}
}

// TestAccKubeOVNNat_basic creates a NAT gateway with an EIP, an SNAT and a
// DNAT rule, checks that the EIP and the rules were ready when the apply
// returned, imports every NAT resource, and destroys everything, which
// covers the order kube-ovn imposes: rules, then EIP, then gateway.
func TestAccKubeOVNNat_basic(t *testing.T) {
	const (
		gatewayResourceName = constants.ResourceTypeKubeOVNVpcNatGateway + ".test"
		eipResourceName     = constants.ResourceTypeKubeOVNIptablesEIP + ".test"
		snatResourceName    = constants.ResourceTypeKubeOVNIptablesSnatRule + ".test"
		dnatResourceName    = constants.ResourceTypeKubeOVNIptablesDnatRule + ".test"
	)
	ctx := context.Background()
	externalSubnet, eipIP := os.Getenv(testAccKubeOVNExternalSubnetEnv), os.Getenv(testAccKubeOVNEIPIPEnv)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccKubeOVNNatPreCheck(ctx, t, externalSubnet, eipIP) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKubeOVNNatDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKubeOVNNatConfig(externalSubnet, eipIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(gatewayResourceName, constants.FieldKubeOVNVpcNatGwLanIP, testAccKubeOVNNatLanIP),
					resource.TestCheckResourceAttr(gatewayResourceName, constants.FieldKubeOVNVpcNatGwExternalSubnets+".0", externalSubnet),
					resource.TestCheckResourceAttr(eipResourceName, constants.FieldCommonState, constants.StateCommonReady),
					resource.TestCheckResourceAttr(eipResourceName, constants.FieldKubeOVNIptablesEIPStatusIP, eipIP),
					resource.TestCheckResourceAttr(snatResourceName, constants.FieldCommonState, constants.StateCommonReady),
					resource.TestCheckResourceAttr(snatResourceName, constants.FieldKubeOVNIptablesSnatStatusCIDR, testAccKubeOVNNatCIDR),
					resource.TestCheckResourceAttr(dnatResourceName, constants.FieldCommonState, constants.StateCommonReady),
					resource.TestCheckResourceAttr(dnatResourceName, constants.FieldKubeOVNIptablesDnatProtocol, constants.KubeOVNIptablesDnatProtocolTCP),
					resource.TestCheckResourceAttr(dnatResourceName, constants.FieldKubeOVNIptablesDnatStatusV4IP, eipIP),
					resource.TestCheckResourceAttrPair("data."+eipResourceName, constants.FieldKubeOVNIptablesEIPStatusIP, eipResourceName, constants.FieldKubeOVNIptablesEIPStatusIP),
				),
			},
			{
				ResourceName:      gatewayResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      eipResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      snatResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      dnatResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
