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
	testAccImageName         = "test-acc-foo"
	testAccImageResourceName = constants.ResourceTypeImage + "." + testAccImageName
	testAccImageDescription  = "Terraform Harvester image acceptance test"
	testAccImageDisplayName  = "foo"
	testAccImageSourceType   = "download"

	testAccImageURL = "http://cloud-images.ubuntu.com/releases/focal/release/ubuntu-20.04-server-cloudimg-amd64.img"

	testAccImageConfigTemplate = `
resource %s "%s" {
	%s = "%s"
	%s = "%s"
	%s = "%s"
	%s = "%s"
	%s = "%s"
}
`
)

func buildImageConfig(name, description, displayName, sourceType, url string) string {
	return fmt.Sprintf(testAccImageConfigTemplate, constants.ResourceTypeImage, name,
		constants.FieldCommonName, name,
		constants.FieldCommonDescription, description,
		constants.FieldImageDisplayName, displayName,
		constants.FieldImageSourceType, sourceType,
		constants.FieldImageURL, url)
}

func TestAccImage_basic(t *testing.T) {
	var (
		image *harvsterv1.VirtualMachineImage
		ctx   = context.Background()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      buildImageConfig(testAccImageName, testAccImageDescription, "", testAccImageSourceType, testAccImageURL),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`%s must not be empty`, constants.FieldImageDisplayName)),
			},
			{
				Config:      buildImageConfig(testAccImageName, testAccImageDescription, testAccImageDisplayName, testAccImageSourceType, ""),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`expected "%s" url to not be empty`, constants.FieldImageURL)),
			},
			{
				Config: buildImageConfig(testAccImageName, testAccImageDescription, testAccImageDisplayName, testAccImageSourceType, testAccImageURL),
				Check: resource.ComposeTestCheckFunc(
					testAccImageExists(ctx, testAccImageResourceName, image),
					resource.TestCheckResourceAttr(testAccImageResourceName, constants.FieldCommonName, testAccImageName),
					resource.TestCheckResourceAttr(testAccImageResourceName, constants.FieldCommonDescription, testAccImageDescription),
					resource.TestCheckResourceAttr(testAccImageResourceName, constants.FieldImageSourceType, testAccImageSourceType),
					resource.TestCheckResourceAttr(testAccImageResourceName, constants.FieldImageURL, testAccImageURL),
				),
			},
		},
	})
}

func testAccImageExists(ctx context.Context, n string, image *harvsterv1.VirtualMachineImage) resource.TestCheckFunc {
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
		foundImage, err := c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		image = foundImage
		return nil
	}
}

func testAccCheckImageDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != constants.ResourceTypeImage {
				continue
			}

			c := testAccProvider.Meta().(*client.Client)
			namespace, name, err := helper.IDParts(rs.Primary.ID)
			if err != nil {
				return err
			}

			imageStateRefreshFunc := getResourceStateRefreshFunc(func() (interface{}, error) {
				return c.HarvesterClient.HarvesterhciV1beta1().VirtualMachineImages(namespace).Get(ctx, name, metav1.GetOptions{})
			})
			stateConf := getStateChangeConf(imageStateRefreshFunc)
			if _, err = stateConf.WaitForStateContext(ctx); err != nil {
				return fmt.Errorf(
					"[ERROR] waiting for image (%s) to be removed: %s", rs.Primary.ID, err)
			}
		}
		return nil
	}
}
