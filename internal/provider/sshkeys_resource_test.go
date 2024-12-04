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
resource "qbee_sshkeys" "test" {
  tag = "terraform:acctest:sshkeys"
  users = [
    {
      username = "testuser"
      userkeys = [
        "key1",
        "key2"
      ]
	}
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_sshkeys.test", "tag", "terraform:acctest:sshkeys"),
					resource.TestCheckNoResourceAttr("qbee_sshkeys.test", "node"),
					resource.TestCheckResourceAttr("qbee_sshkeys.test", "users.#", "1"),
					resource.TestCheckResourceAttr("qbee_sshkeys.test", "users.1.username", "testuser"),
					resource.TestCheckResourceAttr("qbee_sshkeys.test", "users.1.userkeys.#", "2"),
					resource.TestCheckResourceAttr("qbee_sshkeys.test", "users.1.userkeys.1", "key1"),
					resource.TestCheckResourceAttr("qbee_sshkeys.test", "users.1.userkeys.2", "key2"),
				),
			},
			// Update to a different template
			{
				Config: providerConfig + `
resource "qbee_sshkeys" "test" {
  tag = "terraform:acctest:sshkeys"
  users = [
    {
      username = "testuser"
      userkeys = [ "key1" ]
	},
    {
      username = "otheruser"
      userkeys = []
	}
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_sshkeys.test", "tag", "terraform:acctest:sshkeys"),
					resource.TestCheckNoResourceAttr("qbee_sshkeys.test", "node"),
					resource.TestCheckResourceAttr("qbee_sshkeys.test", "users.#", "2"),
					resource.TestCheckResourceAttr("qbee_sshkeys.test", "users.1.username", "testuser"),
					resource.TestCheckResourceAttr("qbee_sshkeys.test", "users.1.userkeys.#", "1"),
					resource.TestCheckResourceAttr("qbee_sshkeys.test", "users.1.userkeys.1", "key1"),
					resource.TestCheckResourceAttr("qbee_sshkeys.test", "users.2.username", "otheruser"),
					resource.TestCheckResourceAttr("qbee_sshkeys.test", "users.2.userkeys.#", "0"),
				),
			},
			// Import from tag
			{
				ResourceName:                         "qbee_sshkeys.test",
				ImportState:                          true,
				ImportStateId:                        "tag:terraform:acctest:sshkeys",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tag",
			},
			// Update to be for a node
			{
				Config: providerConfig + `
resource "qbee_sshkeys" "test" {
  node = "integrationtests"
  users = [
    {
      username = "testuser"
      userkeys = [ "key1" ]
	},
    {
      username = "otheruser"
      userkeys = []
	}
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("qbee_sshkeys.test", "tag"),
					resource.TestCheckResourceAttr("qbee_sshkeys.test", "node", "integrationtests"),
					resource.TestCheckResourceAttr("qbee_sshkeys.test", "users.#", "2"),
					resource.TestCheckResourceAttr("qbee_sshkeys.test", "users.1.username", "testuser"),
					resource.TestCheckResourceAttr("qbee_sshkeys.test", "users.1.userkeys.#", "1"),
					resource.TestCheckResourceAttr("qbee_sshkeys.test", "users.1.userkeys.1", "key1"),
					resource.TestCheckResourceAttr("qbee_sshkeys.test", "users.2.username", "otheruser"),
					resource.TestCheckResourceAttr("qbee_sshkeys.test", "users.2.userkeys.#", "0"),
				),
			},
			// Import from node
			{
				ResourceName:                         "qbee_sshkeys.test",
				ImportState:                          true,
				ImportStateId:                        "node:integrationtests",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "node",
			},
		},
	})
}
