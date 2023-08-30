package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNodeFiledistributionGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "qbee_node_filedistribution" "test" {
  node = "integrationtests"
  extend = true
  files = [
    {
      command = "date -u > /tmp/last-updated.txt"
      pre_condition = "/bin/true"
      templates = [
        {
          source = "/acctest/source"
          destination = "/tmp/target"
          is_template = true
        }
      ]
      parameters = [
        {
          key = "param-key"
          value = "param-value"
        }
      ]
	}
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "id", "placeholder"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "node", "integrationtests"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "files.#", "1"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "files.0.command", "date -u > /tmp/last-updated.txt"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "files.0.pre_condition", "/bin/true"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "files.0.templates.#", "1"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "files.0.templates.0.source", "/acctest/source"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "files.0.templates.0.destination", "/tmp/target"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "files.0.templates.0.is_template", "true"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "files.0.parameters.#", "1"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "files.0.parameters.0.key", "param-key"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "files.0.parameters.0.value", "param-value"),
				),
			},
			// Update
			{
				Config: providerConfig + `
resource "qbee_node_filedistribution" "test" {
  node = "integrationtests"
  extend = true
  files = [
    {
      command = "/bin/true"
      pre_condition = "date -u > /tmp/last-updated.txt"
      templates = [
        {
          source = "/acctest/source2"
          destination = "/tmp/target2"
          is_template = false
        }
      ]
	}
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "id", "placeholder"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "node", "integrationtests"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "files.#", "1"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "files.0.command", "/bin/true"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "files.0.pre_condition", "date -u > /tmp/last-updated.txt"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "files.0.templates.#", "1"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "files.0.templates.0.source", "/acctest/source2"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "files.0.templates.0.destination", "/tmp/target2"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "files.0.templates.0.is_template", "false"),
					resource.TestCheckResourceAttr("qbee_node_filedistribution.test", "files.0.parameters.#", "0"),
				),
			},
			// Import testing
			{
				ResourceName:      "qbee_node_filedistribution.test",
				ImportState:       true,
				ImportStateId:     "integrationtests",
				ImportStateVerify: true,
			},
		},
	})
}
