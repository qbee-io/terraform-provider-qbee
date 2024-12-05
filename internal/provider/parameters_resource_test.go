package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccParametersResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "qbee_parameters" "test" {
  tag = "terraform:acctest:parameters"
  extend = true
  parameters = [
    {
      key = "parameter-key-1"
      value = "parameter-value-1"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_parameters.test", "tag", "terraform:acctest:parameters"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "node"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.#", "1"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.0.key", "parameter-key-1"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.0.value", "parameter-value-1"),
				),
			},
			// Update to a different template
			{
				Config: providerConfig + `
resource "qbee_parameters" "test" {
  tag = "terraform:acctest:parameters"
  extend = false
  parameters = [
    {
      key = "parameter-key-1"
      value = "parameter-value-1"
    },
    {
      key = "parameter-key-2"
      value = "parameter-value-2"
    }
  ]
  secrets = [
    {
      key = "secret-key"
      value = "secret-value"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_parameters.test", "tag", "terraform:acctest:parameters"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "node"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "extend", "false"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.#", "2"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.0.key", "parameter-key-1"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.0.value", "parameter-value-1"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.1.key", "parameter-key-2"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.1.value", "parameter-value-2"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "secrets.#", "1"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "secrets.0.key", "secret-key"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "secrets.0.value", "secret-value"),
					resource.TestCheckResourceAttrSet("qbee_parameters.test", "secrets.0.secret_id"),
				),
			},
			// Import tag
			{
				ResourceName:                         "qbee_parameters.test",
				ImportState:                          true,
				ImportStateId:                        "tag:terraform:acctest:parameters",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tag",
				ImportStateVerifyIgnore: []string{
					"secrets",
				},
			},
			// Update to be for a node
			{
				Config: providerConfig + `
resource "qbee_parameters" "test" {
  node = "integrationtests"
  extend = true
  parameters = [
    {
      key = "parameter-key-1"
      value = "parameter-value-1"
    },
    {
      key = "parameter-key-2"
      value = "parameter-value-2"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "tag"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "node", "integrationtests"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.#", "2"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.0.key", "parameter-key-1"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.0.value", "parameter-value-1"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.1.key", "parameter-key-2"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.1.value", "parameter-value-2"),
				),
			},
			// Import testing
			{
				ResourceName:                         "qbee_parameters.test",
				ImportState:                          true,
				ImportStateId:                        "node:integrationtests",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "node",
			},
		},
	})
}
