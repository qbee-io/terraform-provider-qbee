package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPackageManagementResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "qbee_package_management" "test" {
  tag = "terraform:acctest:packagemanagement"
  extend = true
  full_upgrade = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_package_management.test", "tag", "terraform:acctest:packagemanagement"),
					resource.TestCheckNoResourceAttr("qbee_package_management.test", "node"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "full_upgrade", "true"),
					resource.TestCheckNoResourceAttr("qbee_package_management.test", "pre_condition"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "reboot_mode", "never"),
					resource.TestCheckNoResourceAttr("qbee_package_management.test", "packages"),
				),
			},
			// Update to a different template
			{
				Config: providerConfig + `
resource "qbee_package_management" "test" {
  tag = "terraform:acctest:packagemanagement"
  extend = true
  pre_condition = "true"
  reboot_mode = "never"
  packages = [
	{
	  "name": "vim",
	  "version": "9.1",
	}
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_package_management.test", "tag", "terraform:acctest:packagemanagement"),
					resource.TestCheckNoResourceAttr("qbee_package_management.test", "node"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "pre_condition", "true"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "reboot_mode", "never"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "packages.#", "1"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "packages.0.name", "vim"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "packages.0.version", "9.1"),
				),
			},
			// Import tag
			{
				ResourceName:                         "qbee_package_management.test",
				ImportState:                          true,
				ImportStateId:                        "tag:terraform:acctest:packagemanagement",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tag",
			},
			// Update to be for a node
			{
				Config: providerConfig + `
resource "qbee_package_management" "test" {
  node = "integrationtests"
  extend = true
  reboot_mode = "always"
  packages = [
	{
	  "name": "vim",
	  "version": "9.1",
	}
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("qbee_package_management.test", "tag"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "node", "integrationtests"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "reboot_mode", "always"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "packages.#", "1"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "packages.0.name", "vim"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "packages.0.version", "9.1"),
				),
			},
			// Import testing
			{
				ResourceName:                         "qbee_package_management.test",
				ImportState:                          true,
				ImportStateId:                        "node:integrationtests",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "node",
			},
		},
	})
}
