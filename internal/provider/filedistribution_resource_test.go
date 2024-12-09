package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFiledistributionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "qbee_filedistribution" "test" {
  tag = "terraform:acctest:filedistribution"
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
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "tag", "terraform:acctest:filedistribution"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.#", "1"),
					resource.TestCheckNoResourceAttr("qbee_filedistribution.test", "files.0.label"),
					resource.TestCheckNoResourceAttr("qbee_filedistribution.test", "files.0.command"),
					resource.TestCheckNoResourceAttr("qbee_filedistribution.test", "files.0.pre_condition"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.templates.#", "1"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.templates.0.source", "/acctest/source"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.templates.0.destination", "/tmp/target"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.templates.0.is_template", "false"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.parameters.#", "0"),
				),
			},
			// Update with all fields
			{
				Config: providerConfig + `
resource "qbee_filedistribution" "test" {
  tag = "terraform:acctest:filedistribution"
  extend = true
  files = [
    {
	  label = "acc-test file distribution"
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
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "tag", "terraform:acctest:filedistribution"),
					resource.TestCheckNoResourceAttr("qbee_filedistribution.test", "node"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.#", "1"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.label", "acc-test file distribution"),
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
			// Import from tag
			{
				ResourceName:                         "qbee_filedistribution.test",
				ImportState:                          true,
				ImportStateId:                        "tag:terraform:acctest:filedistribution",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tag",
			},
			// Update to be for a node
			{
				Config: providerConfig + `
resource "qbee_filedistribution" "test" {
  node = "integrationtests"
  extend = true
  files = [
    {
	  label = "acc-test file distribution"
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
					resource.TestCheckNoResourceAttr("qbee_filedistribution.test", "tag"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "node", "integrationtests"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.#", "1"),
					resource.TestCheckResourceAttr("qbee_filedistribution.test", "files.0.label", "acc-test file distribution"),
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
			// Import from node
			{
				ResourceName:                         "qbee_filedistribution.test",
				ImportState:                          true,
				ImportStateId:                        "node:integrationtests",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "node",
			},
		},
	})
}
