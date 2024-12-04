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
  tag = "terraform:acctest:package_management"
  full_upgrade = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_package_management.test", "id", "placeholder"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "tag", "terraform:acctest:package_management"),
					resource.TestCheckNoResourceAttr("qbee_package_management.test", "node"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "full_upgrade", "true"),
					resource.TestCheckNoResourceAttr("qbee_package_management.test", "pre_condition"),
					resource.TestCheckNoResourceAttr("qbee_package_management.test", "reboot_mode"),
					resource.TestCheckNoResourceAttr("qbee_package_management.test", "items"),
				),
			},
			// Update to a different template
			{
				Config: providerConfig + `
resource "qbee_package_management" "test" {
  tag = "terraform:acctest:package_management"
  pre_condition = "true"
  reboot_mode = "never"
  items = [
	{
	  "package": "vim",
	  "version": "9.1",
	}
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_package_management.test", "id", "placeholder"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "tag", "terraform:acctest:package_management"),
					resource.TestCheckNoResourceAttr("qbee_package_management.test", "node"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "pre_condition", "true"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "reboot_mode", "never"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "items.#", "1"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "items.0.package", "vim"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "items.0.version", "9.1"),
				),
			},
			// Import tag
			{
				ResourceName:      "qbee_package_management.test",
				ImportState:       true,
				ImportStateId:     "tag:terraform:acctest:package_management",
				ImportStateVerify: true,
			},
			// Update to be for a node
			{
				Config: providerConfig + `
resource "qbee_package_management" "test" {
  node = "integrationtests"
  reboot_mode = "always"
  items = [
	{
	  "package": "vim",
	  "version": "9.1",
	}
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_package_management.test", "id", "placeholder"),
					resource.TestCheckNoResourceAttr("qbee_package_management.test", "tag"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "node", "integrationtests"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "reboot_mode", "always"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "items.#", "1"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "items.0.package", "vim"),
					resource.TestCheckResourceAttr("qbee_package_management.test", "items.0.version", "9.1"),
				),
			},
			// Import testing
			{
				ResourceName:      "qbee_package_management.test",
				ImportState:       true,
				ImportStateId:     "node:integrationtests",
				ImportStateVerify: true,
			},
		},
	})
}
