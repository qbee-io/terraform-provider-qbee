package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSettingsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "qbee_settings" "test" {
  tag = "terraform:acctest:settings"
  extend = true
  metrics = true
  reports = true
  remote_console = false
  software_inventory = true
  process_inventory = true
  agent_interval = 5
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_settings.test", "tag", "terraform:acctest:settings"),
					resource.TestCheckNoResourceAttr("qbee_settings.test", "node"),
					resource.TestCheckResourceAttr("qbee_settings.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_settings.test", "metrics", "true"),
					resource.TestCheckResourceAttr("qbee_settings.test", "reports", "true"),
					resource.TestCheckResourceAttr("qbee_settings.test", "remote_console", "false"),
					resource.TestCheckResourceAttr("qbee_settings.test", "software_inventory", "true"),
					resource.TestCheckResourceAttr("qbee_settings.test", "process_inventory", "true"),
					resource.TestCheckResourceAttr("qbee_settings.test", "agent_interval", "5"),
				),
			},
			// Update to a different template
			{
				Config: providerConfig + `
resource "qbee_settings" "test" {
  tag = "terraform:acctest:settings"
  extend = false
  metrics = true
  reports = true
  remote_console = false
  software_inventory = true
  process_inventory = true
  agent_interval = 10
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_settings.test", "tag", "terraform:acctest:settings"),
					resource.TestCheckNoResourceAttr("qbee_settings.test", "node"),
					resource.TestCheckResourceAttr("qbee_settings.test", "extend", "false"),
					resource.TestCheckResourceAttr("qbee_settings.test", "metrics", "true"),
					resource.TestCheckResourceAttr("qbee_settings.test", "reports", "true"),
					resource.TestCheckResourceAttr("qbee_settings.test", "remote_console", "false"),
					resource.TestCheckResourceAttr("qbee_settings.test", "software_inventory", "true"),
					resource.TestCheckResourceAttr("qbee_settings.test", "process_inventory", "true"),
					resource.TestCheckResourceAttr("qbee_settings.test", "agent_interval", "10"),
				),
			},
			// Import tag
			{
				ResourceName:                         "qbee_settings.test",
				ImportState:                          true,
				ImportStateId:                        "tag:terraform:acctest:settings",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tag",
			},
			// Update to be for a node
			{
				Config: providerConfig + `
resource "qbee_settings" "test" {
  node = "integrationtests"
  extend = true
  metrics = true
  reports = true
  remote_console = true
  software_inventory = true
  process_inventory = true
  agent_interval = 10
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("qbee_settings.test", "tag"),
					resource.TestCheckResourceAttr("qbee_settings.test", "node", "integrationtests"),
					resource.TestCheckResourceAttr("qbee_settings.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_settings.test", "metrics", "true"),
					resource.TestCheckResourceAttr("qbee_settings.test", "reports", "true"),
					resource.TestCheckResourceAttr("qbee_settings.test", "remote_console", "true"),
					resource.TestCheckResourceAttr("qbee_settings.test", "software_inventory", "true"),
					resource.TestCheckResourceAttr("qbee_settings.test", "process_inventory", "true"),
					resource.TestCheckResourceAttr("qbee_settings.test", "agent_interval", "10"),
				),
			},
			// Import node
			{
				ResourceName:                         "qbee_settings.test",
				ImportState:                          true,
				ImportStateId:                        "node:integrationtests",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "node",
			},
		},
	})
}
