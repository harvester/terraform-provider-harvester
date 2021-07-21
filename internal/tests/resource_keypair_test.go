package tests

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

const (
	testAccKeyPairName         = "test-acc-foo"
	testAccKeyPairResourceName = constants.ResourceTypeKeyPair + "." + testAccKeyPairName
	testAccKeyPairDescription  = "Terraform Harvester keyPair acceptance test"

	testAccKeyPairPublicKey       = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDtlMZYPhQCRU+PgYe8c+YAQghMi7cJpj60HpIGZKgPQ9uLWAIWm6e38mhzNXtbPU4OiFgML9tTJ0UoHzZYueeo1H58R7O65rcQRSCbICj645B/x4MYJhrLMpSLIo1SAT1RiWTaAnffwYPfbQ8r8R1YxizoPZGe4dj4BHwBmEAQhUs4ymxWs5TNmATv/XcF2FnUJvUNATk4WGX+wBNDp6T47wgnMz/D9WOcum776SDozXvfbcZDDpdNNjZyzq3e/admx2+wIatjCFHTQmwgpyphMeOeYKT/w1YitHFlaipys2L4ABQv2gzHjpGVa6YE9B25Ez9vly/hbjz3LN28gf9oBGujWeisU9QZ1Gn97Ao6rF/zStE1c8Glphg60DuYa1JrlbrhmHtMmR9VZ+xROuc//2MixeBPd9ysFoY0Vki+GNLc/gzAGbN0GF6uviFx57pWQ50zoSA9en9Ss9Gji2eYd2yVOSCR4y6NsP/bzDUefnGe87fFfvMAu6LjXIjLwcM="
	testAccKeyPairPublicKeyUpdate = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC8gVNwltyQ/aHbZATnp1lcD4mLHyRjpF3AozA6wcqidDjQ3sda3OIqLLwr1hGcxc5g21mkyt9H6jyoEDnbiYgUfXdmtjCxRUHoHqQYq401ApnF3Nt3KF+aIwtHCJlCIPCRieEBgAc7N/MP6w9fFD1/eJIpUGN50jAH/tVCKq7vlHSATYXsKyhHqv3KJztBctskLDVXoDNaeV9szkGr/fwjTWIgkGZ3YsI74sv+MRcc1ainmQk6Ji89cIml7EOMQJdX38ieUyhXX6vwoa9FnjSysYn/pQZVgyNCy80b3ZkVQPaCslNQGXHactl6SB+S3eiu5gY4LzSW1fzJhN5UT9MUbZLqXowbi71Z453qQbS+FCAch7Pnj7QbW93dN/m8HG1E3OyLpX5q4xtB2OTYatYMO9NQy6z5loCYBuciJGl/sfTci6YFhvyuLA/qpvWmKN3neTMhcglyN+0JmnVSJ39pcCoM3vH4wHdQupeiAl/laIyMN/vBwbH49bTbT3WFT9c="

	testAccKeyPairConfigTemplate = `
resource %s "%s" {
	%s = "%s"
	%s = "%s"
	%s = "%s"
}
`
)

func buildKeyPairConfig(name, description, publicKey string) string {
	return fmt.Sprintf(testAccKeyPairConfigTemplate, constants.ResourceTypeKeyPair, name,
		constants.FieldCommonName, name,
		constants.FieldCommonDescription, description,
		constants.FieldKeyPairPublicKey, publicKey)
}

func TestAccKeyPair_basic(t *testing.T) {
	var (
		keyPair *harvsterv1.KeyPair
		ctx     = context.Background()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKeyPairDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      buildKeyPairConfig(testAccKeyPairName, testAccKeyPairDescription, ""),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`%s must not be empty`, constants.FieldKeyPairPublicKey)),
			},
			{
				Config: buildKeyPairConfig(testAccKeyPairName, testAccKeyPairDescription, testAccKeyPairPublicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccKeyPairExists(ctx, testAccKeyPairResourceName, keyPair),
					resource.TestCheckResourceAttr(testAccKeyPairResourceName, constants.FieldCommonName, testAccKeyPairName),
					resource.TestCheckResourceAttr(testAccKeyPairResourceName, constants.FieldCommonDescription, testAccKeyPairDescription),
					resource.TestCheckResourceAttr(testAccKeyPairResourceName, constants.FieldKeyPairPublicKey, testAccKeyPairPublicKey),
				),
			},
			{
				Config: buildKeyPairConfig(testAccKeyPairName, testAccKeyPairDescription, testAccKeyPairPublicKeyUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccKeyPairExists(ctx, testAccKeyPairResourceName, keyPair),
					resource.TestCheckResourceAttr(testAccKeyPairResourceName, constants.FieldCommonName, testAccKeyPairName),
					resource.TestCheckResourceAttr(testAccKeyPairResourceName, constants.FieldCommonDescription, testAccKeyPairDescription),
					resource.TestCheckResourceAttr(testAccKeyPairResourceName, constants.FieldKeyPairPublicKey, testAccKeyPairPublicKeyUpdate),
				),
			},
		},
	})
}

func testAccKeyPairExists(ctx context.Context, n string, keyPair *harvsterv1.KeyPair) resource.TestCheckFunc {
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
		foundKeyPair, err := c.HarvesterClient.HarvesterhciV1beta1().KeyPairs(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		keyPair = foundKeyPair
		return nil
	}
}

func testAccCheckKeyPairDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != constants.ResourceTypeKeyPair {
				continue
			}

			c := testAccProvider.Meta().(*client.Client)
			namespace, name, err := helper.IDParts(rs.Primary.ID)
			if err != nil {
				return err
			}

			keyPairStateRefreshFunc := getResourceStateRefreshFunc(func() (interface{}, error) {
				return c.HarvesterClient.HarvesterhciV1beta1().KeyPairs(namespace).Get(ctx, name, metav1.GetOptions{})
			})
			stateConf := getStateChangeConf(keyPairStateRefreshFunc)
			if _, err = stateConf.WaitForStateContext(ctx); err != nil {
				return fmt.Errorf(
					"[ERROR] waiting for keyPair (%s) to be removed: %s", rs.Primary.ID, err)
			}
		}
		return nil
	}
}
