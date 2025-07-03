package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

const (
	testAccSecretName         = "test-secret"
	testAccSecretResourceName = constants.ResourceTypeSecret + "." + testAccSecretName
	testAccSecretDescription  = "Terraform Harvester secret acceptance test" // #nosec G101 - test constant, not real credentials
	testAccSecretNamespace    = "default"

	// #nosec G101 - test template, not real credentials
	testAccSecretConfigTemplate = `
resource %s "%s" {
	%s = "%s"
	%s = "%s"
	%s = "%s"
	%s = "%s"
	%s = {
		%s = "%s"
		%s = "%s"
	}
}
`

	// #nosec G101 - test template, not real credentials
	testAccSecretConfigDataTemplate = `
resource %s "%s" {
	%s = "%s"
	%s = "%s"
	%s = "%s"
	%s = "%s"
	%s = {
		%s = "%s"
		%s = "%s"
	}
}
`
)

var (
	// Test data variables to avoid hardcoding in templates
	testSecretKey1       = "test_key"
	testSecretValue1     = "test_value"
	testSecretKey2       = "config_data"           // #nosec G101 - test template, not real credentials
	testSecretValue2     = "sample_config_content" // #nosec G101 - test template, not real credentials
	testSecretDataKey1   = "test_binary"
	testSecretDataValue1 = "dGVzdF9kYXRh" // #nosec G101 - test template, not real credentials, base64 encoded "test_data"
	testSecretDataKey2   = "config_file"
	testSecretDataValue2 = "Y29uZmlnX2NvbnRlbnQ=" // #nosec G101 - test template, not real credentials, base64 encoded "config_content"
)

func TestAccResourceSecret(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccSecretConfigTemplate,
					constants.ResourceTypeSecret, testAccSecretName,
					constants.FieldCommonName, testAccSecretName,
					constants.FieldCommonNamespace, testAccSecretNamespace,
					constants.FieldCommonDescription, testAccSecretDescription,
					constants.FieldSecretType, "Opaque",
					constants.FieldSecretStringData,
					testSecretKey1, testSecretValue1,
					testSecretKey2, testSecretValue2,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(testAccSecretResourceName),
					resource.TestCheckResourceAttr(testAccSecretResourceName, constants.FieldCommonName, testAccSecretName),
					resource.TestCheckResourceAttr(testAccSecretResourceName, constants.FieldCommonNamespace, testAccSecretNamespace),
					resource.TestCheckResourceAttr(testAccSecretResourceName, constants.FieldCommonDescription, testAccSecretDescription),
					resource.TestCheckResourceAttr(testAccSecretResourceName, constants.FieldSecretType, "Opaque"),
					resource.TestCheckResourceAttr(testAccSecretResourceName, constants.FieldSecretStringData+"."+testSecretKey1, testSecretValue1),
					resource.TestCheckResourceAttr(testAccSecretResourceName, constants.FieldSecretStringData+"."+testSecretKey2, testSecretValue2),
				),
			},
			{
				Config: fmt.Sprintf(testAccSecretConfigDataTemplate,
					constants.ResourceTypeSecret, testAccSecretName,
					constants.FieldCommonName, testAccSecretName,
					constants.FieldCommonNamespace, testAccSecretNamespace,
					constants.FieldCommonDescription, testAccSecretDescription,
					constants.FieldSecretType, "Opaque",
					constants.FieldSecretData,
					testSecretDataKey1, testSecretDataValue1,
					testSecretDataKey2, testSecretDataValue2,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(testAccSecretResourceName),
					resource.TestCheckResourceAttr(testAccSecretResourceName, constants.FieldCommonName, testAccSecretName),
					resource.TestCheckResourceAttr(testAccSecretResourceName, constants.FieldCommonNamespace, testAccSecretNamespace),
					resource.TestCheckResourceAttr(testAccSecretResourceName, constants.FieldCommonDescription, testAccSecretDescription),
					resource.TestCheckResourceAttr(testAccSecretResourceName, constants.FieldSecretType, "Opaque"),
					resource.TestCheckResourceAttr(testAccSecretResourceName, constants.FieldSecretData+"."+testSecretDataKey1, testSecretDataValue1),
					resource.TestCheckResourceAttr(testAccSecretResourceName, constants.FieldSecretData+"."+testSecretDataKey2, testSecretDataValue2),
				),
			},
		},
	})
}

func testAccCheckSecretExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource %s not found", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource %s has no ID set", n)
		}

		c := testAccProvider.Meta().(*client.Client)
		namespace, name, err := helper.IDParts(rs.Primary.ID)
		if err != nil {
			return err
		}

		foundSecret, err := c.KubeClient.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if foundSecret.Name != name {
			return fmt.Errorf("secret not found")
		}

		return nil
	}
}

func testAccCheckSecretDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*client.Client)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != constants.ResourceTypeSecret {
			continue
		}

		namespace, name, err := helper.IDParts(rs.Primary.ID)
		if err != nil {
			return err
		}

		foundSecret, err := c.KubeClient.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return nil
		}

		if foundSecret.Name == name {
			return fmt.Errorf("secret still exists")
		}
	}

	return nil
}
