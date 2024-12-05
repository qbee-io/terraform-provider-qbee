package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSSHkeysResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "qbee_ssh_keys" "test" {
  tag = "terraform:acctest:sshkeys"
  extend = true
  users = [
    {
      username = "testuser"
      keys = [
        "key1",
        "key2"
      ]
	}
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "tag", "terraform:acctest:sshkeys"),
					resource.TestCheckNoResourceAttr("qbee_ssh_keys.test", "node"),
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "users.#", "1"),
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "users.0.username", "testuser"),
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "users.0.keys.#", "2"),
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "users.0.keys.0", "key1"),
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "users.0.keys.1", "key2"),
				),
			},
			// Update to a different template
			{
				Config: providerConfig + `
resource "qbee_ssh_keys" "test" {
  tag = "terraform:acctest:sshkeys"
  extend = false
  users = [
    {
      username = "testuser"
      keys = [ "key1" ]
	},
    {
      username = "otheruser"
      keys = []
	}
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "tag", "terraform:acctest:sshkeys"),
					resource.TestCheckNoResourceAttr("qbee_ssh_keys.test", "node"),
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "extend", "false"),
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "users.#", "2"),
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "users.0.username", "testuser"),
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "users.0.keys.#", "1"),
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "users.0.keys.0", "key1"),
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "users.1.username", "otheruser"),
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "users.1.keys.#", "0"),
				),
			},
			// Import from tag
			{
				ResourceName:                         "qbee_ssh_keys.test",
				ImportState:                          true,
				ImportStateId:                        "tag:terraform:acctest:sshkeys",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tag",
			},
			// Update to be for a node
			{
				Config: providerConfig + `
resource "qbee_ssh_keys" "test" {
  node = "integrationtests"
  extend = true
  users = [
    {
      username = "testuser"
      keys = [ "key1" ]
	},
    {
      username = "otheruser"
      keys = []
	}
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("qbee_ssh_keys.test", "tag"),
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "node", "integrationtests"),
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "users.#", "2"),
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "users.0.username", "testuser"),
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "users.0.keys.#", "1"),
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "users.0.keys.0", "key1"),
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "users.1.username", "otheruser"),
					resource.TestCheckResourceAttr("qbee_ssh_keys.test", "users.1.keys.#", "0"),
				),
			},
			// Import from node
			{
				ResourceName:                         "qbee_ssh_keys.test",
				ImportState:                          true,
				ImportStateId:                        "node:integrationtests",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "node",
			},
		},
	})
}
