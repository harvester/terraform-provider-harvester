package tests

import (
	"context"
	"fmt"
	"testing"

	harvsterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

const (
	testAccSettingName         = "default-vm-termination-grace-period-seconds"
	testAccSettingResourceName = constants.ResourceTypeSetting + "." + testAccSettingName
	testAccSettingDescription  = "Terraform Harvester volume acceptance test"

	testAccSettingValue = "200"

	testAccSettingConfigTemplate = `
resource %s "%s" {
	%s = "%s"
	%s = "%s"
	%s = "%s"
}
`
)

func buildSettingConfig(name, description, value string) string {
	return fmt.Sprintf(testAccSettingConfigTemplate, constants.ResourceTypeSetting, name,
		constants.FieldCommonName, name,
		constants.FieldCommonDescription, description,
		constants.FieldSettingValue, value)
}

func TestAccSetting_basic(t *testing.T) {
	var (
		setting *harvsterv1.Setting
		ctx     = context.Background()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSettingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: buildSettingConfig(testAccSettingName, testAccSettingDescription, testAccSettingValue),
				Check: resource.ComposeTestCheckFunc(
					testAccSettingExists(ctx, testAccSettingResourceName, setting, testAccSettingValue),
					resource.TestCheckResourceAttr(testAccSettingResourceName, constants.FieldCommonName, testAccSettingName),
					resource.TestCheckResourceAttr(testAccSettingResourceName, constants.FieldCommonDescription, testAccSettingDescription),
					resource.TestCheckResourceAttr(testAccSettingResourceName, constants.FieldSettingValue, testAccSettingValue),
				),
			},
			{
				Config:  buildSettingConfig(testAccSettingName, testAccSettingDescription, testAccSettingValue),
				Destroy: true,
				Check: resource.ComposeTestCheckFunc(
					testAccSettingExists(ctx, testAccSettingResourceName, setting, ""),
				),
			},
		},
	})
}

func testAccSettingExists(ctx context.Context, n string, setting *harvsterv1.Setting, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Resource %s not found. ", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource %s ID not set. ", n)
		}

		id := rs.Primary.ID
		c, err := testAccProvider.Meta().(*config.Config).K8sClient()
		if err != nil {
			return err
		}

		_, name, err := helper.IDParts(id)
		if err != nil {
			return err
		}
		foundsetting, err := c.HarvesterClient.HarvesterhciV1beta1().Settings().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if foundsetting.Value != value {
			return fmt.Errorf("expected setting value %s, got %s", value, foundsetting.Value)
		}

		setting = foundsetting
		return nil
	}
}

func testAccCheckSettingDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != constants.ResourceTypeVolume {
				continue
			}

			c, err := testAccProvider.Meta().(*config.Config).K8sClient()
			if err != nil {
				return err
			}
			_, name, err := helper.IDParts(rs.Primary.ID)
			if err != nil {
				return err
			}

			volumeStateRefreshFunc := func() (interface{}, string, error) {
				setting, err := c.HarvesterClient.HarvesterhciV1beta1().Settings().Get(ctx, name, metav1.GetOptions{})
				if err != nil {
					return nil, testAccResourceStateError, err
				}
				if setting.Value == "" {
					return setting, testAccResourceStateRemoved, nil
				}
				return setting, testAccResourceStateExist, nil
			}
			stateConf := getStateChangeConf(volumeStateRefreshFunc)
			if _, err = stateConf.WaitForStateContext(ctx); err != nil {
				return fmt.Errorf(
					"[ERROR] waiting for setting (%s) to be reset: %s", rs.Primary.ID, err)
			}
		}
		return nil
	}
}
