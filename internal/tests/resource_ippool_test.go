package tests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	// loadbalancerv1 "github.com/harvester/harvester-load-balancer/pkg/apis/loadbalancer.harvesterhci.io/v1beta1"
)

func TestAccIPPool_basic(t *testing.T) {
	// var (
	// 	ippool *loadbalancerv1.IPPool
	// 	ctx = context.Background()
	// )

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
