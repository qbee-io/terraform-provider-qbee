package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUsersResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "qbee_users" "test" {
  tag = "terraform:acctest:users"
  extend = true
  users = [
    {
      username = "test-user-1"
      action = "add"
	}
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_users.test", "tag", "terraform:acctest:users"),
					resource.TestCheckNoResourceAttr("qbee_users.test", "node"),
					resource.TestCheckResourceAttr("qbee_users.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_users.test", "users.#", "1"),
					resource.TestCheckResourceAttr("qbee_users.test", "users.0.username", "test-user-1"),
					resource.TestCheckResourceAttr("qbee_users.test", "users.0.action", "add"),
				),
			},
			// Update to a different template
			{
				Config: providerConfig + `
resource "qbee_users" "test" {
  tag = "terraform:acctest:users"
  extend = false
  users = [
    {
      username = "test-user-2"
      action = "add"
	},
    {
      username = "default-user"
      action = "remove"
	}
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_users.test", "tag", "terraform:acctest:users"),
					resource.TestCheckNoResourceAttr("qbee_users.test", "node"),
					resource.TestCheckResourceAttr("qbee_users.test", "extend", "false"),
					resource.TestCheckResourceAttr("qbee_users.test", "users.#", "2"),
					resource.TestCheckResourceAttr("qbee_users.test", "users.0.username", "test-user-2"),
					resource.TestCheckResourceAttr("qbee_users.test", "users.0.action", "add"),
					resource.TestCheckResourceAttr("qbee_users.test", "users.1.username", "default-user"),
					resource.TestCheckResourceAttr("qbee_users.test", "users.1.action", "remove"),
				),
			},
			// Import from tag
			{
				ResourceName:                         "qbee_users.test",
				ImportState:                          true,
				ImportStateId:                        "tag:terraform:acctest:users",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tag",
			},
			// Update to be for a node
			{
				Config: providerConfig + `
resource "qbee_users" "test" {
  node = "integrationtests"
  extend = true
  users = [
    {
      username = "test-user-2"
      action = "add"
	},
    {
      username = "default-user"
      action = "remove"
	}
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("qbee_users.test", "tag"),
					resource.TestCheckResourceAttr("qbee_users.test", "node", "integrationtests"),
					resource.TestCheckResourceAttr("qbee_users.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_users.test", "users.#", "2"),
					resource.TestCheckResourceAttr("qbee_users.test", "users.0.username", "test-user-2"),
					resource.TestCheckResourceAttr("qbee_users.test", "users.0.action", "add"),
					resource.TestCheckResourceAttr("qbee_users.test", "users.1.username", "default-user"),
					resource.TestCheckResourceAttr("qbee_users.test", "users.1.action", "remove"),
				),
			},
			// Import from node
			{
				ResourceName:                         "qbee_users.test",
				ImportState:                          true,
				ImportStateId:                        "node:integrationtests",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "node",
			},
		},
	})
}
