package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccConnectivityWatchdogResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "qbee_connectivity_watchdog" "test" {
  tag = "terraform:acctest:connectivity_watchdog"
  extend = false
  threshold = "5"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_connectivity_watchdog.test", "id", "placeholder"),
					resource.TestCheckResourceAttr("qbee_connectivity_watchdog.test", "tag", "terraform:acctest:connectivity_watchdog"),
					resource.TestCheckNoResourceAttr("qbee_connectivity_watchdog.test", "node"),
					resource.TestCheckResourceAttr("qbee_connectivity_watchdog.test", "extend", "false"),
					resource.TestCheckResourceAttr("qbee_connectivity_watchdog.test", "threshold", "5"),
				),
			},
			// Import from tag
			{
				ResourceName:      "qbee_connectivity_watchdog.test",
				ImportState:       true,
				ImportStateId:     "tag:terraform:acctest:connectivity_watchdog",
				ImportStateVerify: true,
			},
			// Update
			{
				Config: providerConfig + `
resource "qbee_connectivity_watchdog" "test" {
  node = "terraform:acctest:connectivity_watchdog"
  extend = true
  threshold = "3"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_connectivity_watchdog.test", "id", "placeholder"),
					resource.TestCheckResourceAttr("qbee_connectivity_watchdog.test", "tag", "terraform:acctest:connectivity_watchdog"),
					resource.TestCheckNoResourceAttr("qbee_connectivity_watchdog.test", "node"),
					resource.TestCheckResourceAttr("qbee_connectivity_watchdog.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_connectivity_watchdog.test", "threshold", "3"),
				),
			},
			// Import testing
			{
				ResourceName:      "qbee_connectivity_watchdog.test",
				ImportState:       true,
				ImportStateId:     "node:integrationtests",
				ImportStateVerify: true,
			},
		},
	})
}
