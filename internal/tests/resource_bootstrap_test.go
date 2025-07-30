package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/harvester/terraform-provider-harvester/internal/provider/bootstrap"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

const (
	testAccBootstrapName            = "test-admin"
	testAccBootstrapResourceName    = constants.ResourceTypeBootstrap + "." + testAccBootstrapName
	testAccBootstrapAPIURL          = "https://192.168.3.131"
	testAccBootstrapInitialPassword = "admin"
	testAccBootstrapPassword        = "passwdpasswd"

	testAccBootstrapConfigTemplate = `
provider "harvester" {
  bootstrap = true
}

resource harvester_bootstrap "%s" {

	%s = "%s"
	%s = "%s"
	%s = "%s"
}
`
)

func buildBootstrapConfig(name, url, initialPassword, password string) string {
	return fmt.Sprintf(testAccBootstrapConfigTemplate, name,
		constants.FieldBootstrapAPIURL, url,
		constants.FieldBootstrapInitialPassword, initialPassword,
		constants.FieldBootstrapPassword, password)
}

func TestAccBootstrap_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: buildBootstrapConfig(testAccBootstrapName, testAccBootstrapAPIURL, testAccBootstrapInitialPassword, testAccBootstrapPassword),
				Check: resource.ComposeTestCheckFunc(
					testAccBootstrapLogin(context.Background(), testAccBootstrapResourceName),
					resource.TestCheckResourceAttr(testAccBootstrapResourceName, constants.FieldBootstrapAPIURL, testAccBootstrapAPIURL),
					resource.TestCheckResourceAttr(testAccBootstrapResourceName, constants.FieldBootstrapInitialPassword, testAccBootstrapInitialPassword),
					resource.TestCheckResourceAttr(testAccBootstrapResourceName, constants.FieldBootstrapPassword, testAccBootstrapPassword),
				),
			},
		},
	})
}

func testAccBootstrapLogin(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if _, ok := s.RootModule().Resources[n]; !ok {
			return fmt.Errorf("Resource %s not found. ", n)
		}

		_, _, err := bootstrap.DoUserLogin(testAccBootstrapAPIURL, "admin", testAccBootstrapPassword, 600, "This is test", "", true)
		return err
	}
}
