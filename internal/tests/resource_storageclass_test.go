package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

const (
	testAccStorageClassName         = "test-acc-foo"
	testAccStorageClassResourceName = constants.ResourceTypeStorageClass + "." + testAccStorageClassName
	testAccStorageClassDescription  = "Terraform Harvester StorageClass acceptance test"

	testAccStorageClassConfigTemplate = `
resource %s "%s" {
	%s = "%s"
	%s = "%s"
	parameters = {
	}
}
`
)

func buildStorageClassConfig(name, description string) string {
	return fmt.Sprintf(testAccStorageClassConfigTemplate, constants.ResourceTypeStorageClass, name,
		constants.FieldCommonDescription, description,
		constants.FieldCommonName, name,
	)
}

func TestAccStorageClass_basic(t *testing.T) {
	var (
		storageClass *storagev1.StorageClass
		ctx          = context.Background()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStorageClassDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: buildStorageClassConfig(testAccStorageClassName, testAccStorageClassDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccStorageClassExists(ctx, testAccStorageClassResourceName, storageClass),
					resource.TestCheckResourceAttr(testAccStorageClassResourceName, constants.FieldCommonName, testAccStorageClassName),
					resource.TestCheckResourceAttr(testAccStorageClassResourceName, constants.FieldCommonDescription, testAccStorageClassDescription),
				),
			},
		},
	})
}

func testAccStorageClassExists(ctx context.Context, n string, storageClass *storagev1.StorageClass) resource.TestCheckFunc {
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

		_, name, err := helper.IDParts(id)
		if err != nil {
			return err
		}
		foundStorageClass, err := c.StorageClassClient.StorageClasses().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		storageClass = foundStorageClass
		return nil
	}
}

func testAccCheckStorageClassDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != constants.ResourceTypeStorageClass {
				continue
			}

			c := testAccProvider.Meta().(*client.Client)
			_, name, err := helper.IDParts(rs.Primary.ID)
			if err != nil {
				return err
			}

			storageClassStateRefreshFunc := getResourceStateRefreshFunc(func() (interface{}, error) {
				return c.StorageClassClient.StorageClasses().Get(ctx, name, metav1.GetOptions{})
			})
			stateConf := getStateChangeConf(storageClassStateRefreshFunc)
			if _, err = stateConf.WaitForStateContext(ctx); err != nil {
				return fmt.Errorf(
					"[ERROR] waiting for storageClass (%s) to be removed: %s", rs.Primary.ID, err)
			}
		}
		return nil
	}
}
