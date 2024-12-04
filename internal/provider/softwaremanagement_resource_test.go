package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSoftwareManagementResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "qbee_softwaremanagement" "test" {
  node = "root"
  extend = true

  items = [
    {
      package = "mysql",
      service_name = "service",
      pre_condition = "/a/b/c",
      config_files = [
        {
          template = "/testtemplate"
          location = "/testlocation"
        }
      ]
      parameters = [
        {
          key = "param1"
          value = "value1"
        }
      ]
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "node", "root"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "items.#", "1"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "items.0.package", "mysql"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "items.0.service_name", "service"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "items.0.pre_condition", "/a/b/c"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "items.0.config_files.#", "1"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "items.0.config_files.0.template", "/testtemplate"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "items.0.config_files.0.location", "/testlocation"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "items.0.parameters.#", "1"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "items.0.parameters.0.key", "param1"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "items.0.parameters.0.value", "value1"),
				),
			},
			// Update testing
			{
				Config: providerConfig + `
resource "qbee_softwaremanagement" "test" {
  tag = "terraform:acctest:softwaremanagement"
  extend = false

  items = [
    {
      package = "mysql",
    },
    {
      package = "customservice",
      service_name = "customservice_name",
      pre_condition = "/bin/true",
      config_files = [
        {
          template = "/testtemplate"
          location = "/testlocation"
        }
      ]
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "tag", "terraform:acctest:softwaremanagement"),
					resource.TestCheckNoResourceAttr("qbee_softwaremanagement.test", "node"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "extend", "false"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "items.#", "2"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "items.0.package", "mysql"),
					resource.TestCheckNoResourceAttr("qbee_softwaremanagement.test", "items.0.service_name"),
					resource.TestCheckNoResourceAttr("qbee_softwaremanagement.test", "items.0.pre_condition"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "items.0.config_files.#", "0"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "items.0.parameters.#", "0"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "items.1.package", "customservice"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "items.1.service_name", "customservice_name"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "items.1.pre_condition", "/bin/true"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "items.1.config_files.#", "1"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "items.1.config_files.0.template", "/testtemplate"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "items.1.config_files.0.location", "/testlocation"),
					resource.TestCheckResourceAttr("qbee_softwaremanagement.test", "items.1.parameters.#", "0"),
				),
			},
			// Import testing
			{
				ResourceName:                         "qbee_softwaremanagement.test",
				ImportState:                          true,
				ImportStateId:                        "tag:terraform:acctest:softwaremanagement",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tag",
			},
		},
	})
}
