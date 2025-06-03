package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBootstrapKeyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "qbee_bootstrap_key" "test" {
  group_id = "terraform:acctest:group"
  auto_accept = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("qbee_bootstrap_key.test", "id"),
					resource.TestCheckResourceAttr("qbee_bootstrap_key.test", "group_id", "terraform:acctest:group"),
					resource.TestCheckResourceAttr("qbee_bootstrap_key.test", "auto_accept", "true"),
				),
			},
			// Update
			{
				Config: providerConfig + `
resource "qbee_bootstrap_key" "test" {
  group_id = "terraform:acctest:other-group"
  auto_accept = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("qbee_bootstrap_key.test", "id"),
					resource.TestCheckResourceAttr("qbee_bootstrap_key.test", "group_id", "terraform:acctest:other-group"),
					resource.TestCheckResourceAttr("qbee_bootstrap_key.test", "auto_accept", "false"),
				),
			},
			// Import testing
			{
				ResourceName:      "qbee_bootstrap_key.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
