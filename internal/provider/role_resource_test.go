package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "qbee_role" "test" {
  name = "terraform:acctest:role"
  description = "Terraform acceptance test role"
  policies = [
	{
	  permission = "device:read"
	  resources = ["*"]
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("qbee_role.test", "id"),
					resource.TestCheckResourceAttr("qbee_role.test", "name", "terraform:acctest:role"),
					resource.TestCheckResourceAttr("qbee_role.test", "description", "Terraform acceptance test role"),
					resource.TestCheckResourceAttr("qbee_role.test", "policies.#", "1"),
					resource.TestCheckResourceAttr("qbee_role.test", "policies.0.permission", "device:read"),
					resource.TestCheckResourceAttr("qbee_role.test", "policies.0.resources.#", "1"),
					resource.TestCheckResourceAttr("qbee_role.test", "policies.0.resources.0", "*"),
				),
			},
			// Update with different policies
			{
				Config: providerConfig + `
resource "qbee_role" "test" {
  name = "terraform:acctest:role"
  description = "Terraform acceptance test role updated"
  policies = [
    {
      permission = "device:read"
	  resources = ["*"]
    },
    {
      permission = "files:read"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_role.test", "name", "terraform:acctest:role"),
					resource.TestCheckResourceAttr("qbee_role.test", "description", "Terraform acceptance test role updated"),
					resource.TestCheckResourceAttr("qbee_role.test", "policies.#", "2"),
					resource.TestCheckResourceAttr("qbee_role.test", "policies.0.permission", "device:read"),
					resource.TestCheckResourceAttr("qbee_role.test", "policies.0.resources.#", "1"),
					resource.TestCheckResourceAttr("qbee_role.test", "policies.0.resources.0", "*"),
					resource.TestCheckResourceAttr("qbee_role.test", "policies.1.permission", "files:read"),
					resource.TestCheckResourceAttr("qbee_role.test", "policies.1.resources.#", "0"),
				),
			},
			// Rename
			{
				Config: providerConfig + `
resource "qbee_role" "test" {
  name = "terraform:acctest:renamed-role"
  description = "Terraform acceptance test role updated"
  policies = [
    {
      permission = "device:read"
	  resources = ["*"]
    },
    {
      permission = "files:read"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_role.test", "name", "terraform:acctest:renamed-role"),
					resource.TestCheckResourceAttr("qbee_role.test", "description", "Terraform acceptance test role updated"),
					resource.TestCheckResourceAttr("qbee_role.test", "policies.#", "2"),
					resource.TestCheckResourceAttr("qbee_role.test", "policies.0.permission", "device:read"),
					resource.TestCheckResourceAttr("qbee_role.test", "policies.0.resources.#", "1"),
					resource.TestCheckResourceAttr("qbee_role.test", "policies.0.resources.0", "*"),
					resource.TestCheckResourceAttr("qbee_role.test", "policies.1.permission", "files:read"),
					resource.TestCheckResourceAttr("qbee_role.test", "policies.1.resources.#", "0"),
				),
			},
			// Import testing
			{
				ResourceName:                         "qbee_role.test",
				ImportState:                          true,
				ImportStateId:                        "terraform:acctest:renamed-role",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
			},
		},
	})
}
