package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFiledistributionGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "qbee_filedistribution" "test" {
  tag = "terraform:acctest:tag"
  extend = true
  files = [
    {
      templates = [
        {
          source = "/acctest/source"
          destination = "/tmp/target"
          is_template = false
        }
      ]
	}
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "id", "placeholder"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "tag", "terraform:acctest:tag"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.#", "1"),
					resource.TestCheckNoResourceAttr("qbee_filedistribution.test", "files.0.command"),
					resource.TestCheckNoResourceAttr("qbee_filedistribution.test", "files.0.pre_condition"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.templates.#", "1"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.templates.0.source", "/acctest/source"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.templates.0.destination", "/tmp/target"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.templates.0.is_template", "false"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.parameters.#", "0"),
				),
			},
			// Update to a different template
			{
				Config: providerConfig + `
resource "qbee_filedistribution" "test" {
  tag = "terraform:acctest:tag"
  extend = true
  files = [
    {
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
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "id", "placeholder"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "tag", "terraform:acctest:tag"),
					resource.TestCheckNoResourceAttr("qbee_filedistribution.test", "node"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.#", "1"),
					resource.TestCheckNoResourceAttr("qbee_filedistribution.test", "files.0.command"),
					resource.TestCheckNoResourceAttr("qbee_filedistribution.test", "files.0.pre_condition"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.templates.#", "1"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.templates.0.source", "/acctest/source2"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.templates.0.destination", "/tmp/target2"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.templates.0.is_template", "false"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.parameters.#", "0"),
				),
			},
			// Update to be for a node
			{
				Config: providerConfig + `
resource "qbee_filedistribution" "test" {
  node = "integrationtests"
  extend = true
  files = [
    {
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
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "id", "placeholder"),
					resource.TestCheckNoResourceAttr("qbee_filedistribution.test", "tag"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "node", "integrationtests"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.#", "1"),
					resource.TestCheckNoResourceAttr("qbee_filedistribution.test", "files.0.command"),
					resource.TestCheckNoResourceAttr("qbee_filedistribution.test", "files.0.pre_condition"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.templates.#", "1"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.templates.0.source", "/acctest/source2"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.templates.0.destination", "/tmp/target2"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.templates.0.is_template", "false"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.parameters.#", "0"),
				),
			},
			// Update again, adding a command, parameters and precondition
			{
				Config: providerConfig + `
resource "qbee_filedistribution" "test" {
  tag = "terraform:acctest:tag"
  extend = true
  files = [
    {
      pre_condition = "/bin/true"
      command = "date -u > /tmp/last-updated.txt"
      templates = [
        {
          source = "/acctest/source2"
          destination = "/tmp/target2"
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
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "id", "placeholder"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "tag", "terraform:acctest:tag"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.#", "1"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.command", "date -u > /tmp/last-updated.txt"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.pre_condition", "/bin/true"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.templates.#", "1"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.templates.0.source", "/acctest/source2"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.templates.0.destination", "/tmp/target2"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.templates.0.is_template", "true"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.parameters.#", "1"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.parameters.0.key", "param-key"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.parameters.0.value", "param-value"),
				),
			},
			// Import testing
			{
				ResourceName:      "qbee_filedistribution.test",
				ImportState:       true,
				ImportStateId:     "tag:terraform:acctest:tag",
				ImportStateVerify: true,
			},
		},
	})
}
