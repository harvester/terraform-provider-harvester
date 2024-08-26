package tests

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	nadv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

const (
	testAccNetworkName         = "test-acc-foo"
	testAccNetworkResourceName = constants.ResourceTypeNetwork + "." + testAccNetworkName
	testAccNetworkDescription  = "Terraform Harvester Network acceptance test"

	testAccNetworkClusterNetworkName = "mgmt"
	testAccNetworkVlanID             = "0"

	testAccNetworkConfigTemplate = `
resource %s "%s" {
	%s = "%s"
	%s = "%s"
	%s = "%s"
	%s = %s
}
`
)

func buildNetworkConfig(name, description, clusterNetworkName, vlanID string) string {
	return fmt.Sprintf(testAccNetworkConfigTemplate, constants.ResourceTypeNetwork, name,
		constants.FieldCommonName, name,
		constants.FieldCommonDescription, description,
		constants.FieldNetworkClusterNetworkName, clusterNetworkName,
		constants.FieldNetworkVlanID, vlanID)
}

func TestAccNetwork_basic(t *testing.T) {
	var (
		network *nadv1.NetworkAttachmentDefinition
		ctx     = context.Background()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: buildNetworkConfig(
					testAccNetworkName,
					testAccNetworkDescription,
					testAccNetworkClusterNetworkName,
					"4095",
				),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`expected %s to be in the range \(0 - 4094\)`, constants.FieldNetworkVlanID)),
			},
			{
				Config: buildNetworkConfig(
					testAccNetworkName,
					testAccNetworkDescription,
					testAccNetworkClusterNetworkName,
					testAccNetworkVlanID,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccNetworkExists(ctx, testAccNetworkResourceName, network),
					resource.TestCheckResourceAttr(testAccNetworkResourceName, constants.FieldCommonName, testAccNetworkName),
					resource.TestCheckResourceAttr(testAccNetworkResourceName, constants.FieldCommonDescription, testAccNetworkDescription),
					resource.TestCheckResourceAttr(testAccNetworkResourceName, constants.FieldNetworkVlanID, testAccNetworkVlanID),
				),
			},
		},
	})
}

func testAccNetworkExists(ctx context.Context, n string, network *nadv1.NetworkAttachmentDefinition) resource.TestCheckFunc {
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
		foundNetwork, err := c.HarvesterClient.K8sCniCncfIoV1().NetworkAttachmentDefinitions(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		network = foundNetwork
		return nil
	}
}

func testAccCheckNetworkDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != constants.ResourceTypeNetwork {
				continue
			}

			c := testAccProvider.Meta().(*client.Client)
			namespace, name, err := helper.IDParts(rs.Primary.ID)
			if err != nil {
				return err
			}

			networkStateRefreshFunc := getResourceStateRefreshFunc(func() (interface{}, error) {
				return c.HarvesterClient.K8sCniCncfIoV1().NetworkAttachmentDefinitions(namespace).Get(ctx, name, metav1.GetOptions{})
			})
			stateConf := getStateChangeConf(networkStateRefreshFunc)
			if _, err = stateConf.WaitForStateContext(ctx); err != nil {
				return fmt.Errorf(
					"[ERROR] waiting for Network (%s) to be removed: %s", rs.Primary.ID, err)
			}
		}
		return nil
	}
}
