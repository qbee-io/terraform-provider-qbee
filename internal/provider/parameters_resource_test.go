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
  parameters = [
    {
      key = "parameter-key-1"
      value = "parameter-value-1"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_parameters.test", "id", "placeholder"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "tag", "terraform:acctest:parameters"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "node"),
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
					resource.TestCheckResourceAttr("qbee_parameters.test", "id", "placeholder"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "tag", "terraform:acctest:parameters"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "node"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.#", "2"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.0.key", "parameter-key-1"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.0.value", "parameter-value-1"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.1.key", "parameter-key-2"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.1.value", "parameter-value-2"),
				),
			},
			// Import tag
			{
				ResourceName:      "qbee_parameters.test",
				ImportState:       true,
				ImportStateId:     "tag:terraform:acctest:parameters",
				ImportStateVerify: true,
			},
			// Update to be for a node
			{
				Config: providerConfig + `
resource "qbee_parameters" "test" {
  node = "integrationtests"
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
					resource.TestCheckResourceAttr("qbee_parameters.test", "id", "placeholder"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "tag"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "node", "integrationtests"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.#", "2"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.0.key", "parameter-key-1"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.0.value", "parameter-value-1"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.1.key", "parameter-key-2"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.1.value", "parameter-value-2"),
				),
			},
			// Import testing
			{
				ResourceName:      "qbee_parameters.test",
				ImportState:       true,
				ImportStateId:     "node:integrationtests",
				ImportStateVerify: true,
			},
		},
	})
}
