package tests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	// loadbalancerv1 "github.com/harvester/harvester-load-balancer/pkg/apis/loadbalancer.harvesterhci.io/v1beta1"
)

func TestIPPoolBasic(t *testing.T) {
	// var (
	// 	ippool *loadbalancerv1.IPPool
	// 	ctx = context.Background()
	// )

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource harvester_ippool \"test_ippool\" {
	name = \"test_ippool\"

	range {
		range_start = \"192.168.0.1\"
		range_end = \"192.168.0.254\"
		range_subnet = \"192.168.0.1/24\"
		range_gateway = \"192.168.0.1\"
	}
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("harvester_ippool.test_ippool", "name", "test_ippool"),
				),
			},
		},
	})
}
