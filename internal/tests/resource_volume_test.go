package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1beta1"

	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

const (
	testAccVolumeName         = "test-acc-foo"
	testAccVolumeResourceName = constants.ResourceTypeVolume + "." + testAccVolumeName
	testAccVolumeDescription  = "Terraform Harvester volume acceptance test"

	testAccVolumeSize = "1Gi"

	testAccVolumeConfigTemplate = `
resource %s "%s" {
	%s = "%s"
	%s = "%s"
	%s = "%s"
}
`
)

func buildVolumeConfig(name, description, size string) string {
	return fmt.Sprintf(testAccVolumeConfigTemplate, constants.ResourceTypeVolume, name,
		constants.FieldCommonName, name,
		constants.FieldCommonDescription, description,
		constants.FieldVolumeSize, size)
}

func TestAccVolume_basic(t *testing.T) {
	var (
		volume *cdiv1beta1.DataVolume
		ctx    = context.Background()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: buildVolumeConfig(testAccVolumeName, testAccVolumeDescription, testAccVolumeSize),
				Check: resource.ComposeTestCheckFunc(
					testAccVolumeExists(ctx, testAccVolumeResourceName, volume),
					resource.TestCheckResourceAttr(testAccVolumeResourceName, constants.FieldCommonName, testAccVolumeName),
					resource.TestCheckResourceAttr(testAccVolumeResourceName, constants.FieldCommonDescription, testAccVolumeDescription),
					resource.TestCheckResourceAttr(testAccVolumeResourceName, constants.FieldVolumeSize, testAccVolumeSize),
				),
			},
		},
	})
}

func testAccVolumeExists(ctx context.Context, n string, volume *cdiv1beta1.DataVolume) resource.TestCheckFunc {
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
		foundVolume, err := c.HarvesterClient.CdiV1beta1().DataVolumes(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		volume = foundVolume
		return nil
	}
}

func testAccCheckVolumeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != constants.ResourceTypeVolume {
				continue
			}

			c := testAccProvider.Meta().(*client.Client)
			namespace, name, err := helper.IDParts(rs.Primary.ID)
			if err != nil {
				return err
			}

			volumeStateRefreshFunc := getResourceStateRefreshFunc(func() (interface{}, error) {
				return c.HarvesterClient.CdiV1beta1().DataVolumes(namespace).Get(ctx, name, metav1.GetOptions{})
			})
			stateConf := getStateChangeConf(volumeStateRefreshFunc)
			if _, err = stateConf.WaitForStateContext(ctx); err != nil {
				return fmt.Errorf(
					"[ERROR] waiting for volume (%s) to be removed: %s", rs.Primary.ID, err)
			}
		}
		return nil
	}
}
