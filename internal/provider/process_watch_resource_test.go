package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProcessWatchResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "qbee_process_watch" "test" {
  tag = "terraform:acctest:processwatch"
  extend = true
  processes = [
    {
      name = "presentProcess"
      policy = "Present"
      command = "start.sh"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_process_watch.test", "tag", "terraform:acctest:processwatch"),
					resource.TestCheckNoResourceAttr("qbee_process_watch.test", "node"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "processes.#", "1"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "processes.0.name", "presentProcess"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "processes.0.policy", "Present"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "processes.0.command", "start.sh"),
				),
			},
			// Update to a different template
			{
				Config: providerConfig + `
resource "qbee_process_watch" "test" {
  tag = "terraform:acctest:processwatch"
  extend = false
  processes = [
    {
      name = "presentProcess"
      policy = "Present"
      command = "start.sh"
    },
    {
	  name = "absentProcess"
	  policy = "Absent"
	  command = "stop.sh"
	}
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_process_watch.test", "tag", "terraform:acctest:processwatch"),
					resource.TestCheckNoResourceAttr("qbee_process_watch.test", "node"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "extend", "false"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "processes.#", "2"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "processes.0.name", "presentProcess"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "processes.0.policy", "Present"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "processes.0.command", "start.sh"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "processes.1.name", "absentProcess"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "processes.1.policy", "Absent"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "processes.1.command", "stop.sh"),
				),
			},
			// Import tag
			{
				ResourceName:                         "qbee_process_watch.test",
				ImportState:                          true,
				ImportStateId:                        "tag:terraform:acctest:processwatch",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tag",
			},
			// Update to be for a node
			{
				Config: providerConfig + `
resource "qbee_process_watch" "test" {
  node = "integrationtests"
  extend = true
  processes = [
    {
      name = "presentProcess"
      policy = "Present"
      command = "start.sh"
    },
    {
	  name = "absentProcess"
	  policy = "Absent"
	  command = "stop.sh"
	}
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("qbee_process_watch.test", "tag"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "node", "integrationtests"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "processes.#", "2"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "processes.0.name", "presentProcess"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "processes.0.policy", "Present"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "processes.0.command", "start.sh"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "processes.1.name", "absentProcess"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "processes.1.policy", "Absent"),
					resource.TestCheckResourceAttr("qbee_process_watch.test", "processes.1.command", "stop.sh"),
				),
			},
			// Import testing
			{
				ResourceName:                         "qbee_process_watch.test",
				ImportState:                          true,
				ImportStateId:                        "node:integrationtests",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "node",
			},
		},
	})
}
