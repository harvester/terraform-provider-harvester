package tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIPPool_invalid(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "harvester_ippool" "test_ippool" {
}
`,
				ExpectError: regexp.MustCompile(`The argument "name" is required, but no definition was found.`),
			},
			{
				Config: `
resource "harvester_ippool" "test_ippool" {
	name = "test-ippool"
}
`,
				ExpectError: regexp.MustCompile(`The argument "range" is required, but no definition was found.`),
			},
		},
	})
}

func TestAccIPPool_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "harvester_ippool" "test_ippool" {
	name = "test-ippool"

	range {
		start = "192.168.0.1"
		end = "192.168.0.254"
		subnet = "192.168.0.1/24"
		gateway = "192.168.0.1"
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("harvester_ippool.test_ippool", "name", "test-ippool"),
					resource.TestCheckResourceAttr("harvester_ippool.test_ippool", "range.#", "1"),
					resource.TestCheckResourceAttr("harvester_ippool.test_ippool", "range.0.start", "192.168.0.1"),
					resource.TestCheckResourceAttr("harvester_ippool.test_ippool", "range.0.end", "192.168.0.254"),
					resource.TestCheckResourceAttr("harvester_ippool.test_ippool", "range.0.subnet", "192.168.0.1/24"),
					resource.TestCheckResourceAttr("harvester_ippool.test_ippool", "range.0.gateway", "192.168.0.1"),
				),
			},
		},
	})
}
