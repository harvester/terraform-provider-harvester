package tests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	// loadbalancerv1 "github.com/harvester/harvester-load-balancer/pkg/apis/loadbalancer.harvesterhci.io/v1beta1"
)

func TestLoadBalancerBasic(t *testing.T) {
	// var (
	// 	loadbalancer *loadbalancerv1.LoadBalancer
	// 	ctx = context.Background()
	// )

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "harvester_loadbalancer" "test_loadbalancer" {
	name = "test-loadbalancer"

	listener {
		port = 443
		protocol = "tcp"
		backend_port = 8080
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("harvester_loadbalancer.test_loadbalancer", "name", "test-loadbalancer"),
					resource.TestCheckResourceAttr("harvester_loadbalancer.test_loadbalancer", "listener.#", "1"),
					resource.TestCheckResourceAttr("harvester_loadbalancer.test_loadbalancer", "listener.0.port", "443"),
					resource.TestCheckResourceAttr("harvester_loadbalancer.test_loadbalancer", "listener.0.protocol", "tcp"),
					resource.TestCheckResourceAttr("harvester_loadbalancer.test_loadbalancer", "listener.0.backend_port", "8080"),
				),
			},
		},
	})
}
