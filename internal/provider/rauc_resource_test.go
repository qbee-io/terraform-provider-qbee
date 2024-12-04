package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRaucResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "qbee_rauc" "test" {
  tag = "terraform:acctest:rauc"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_rauc.test", "tag", "terraform:acctest:rauc"),
					resource.TestCheckNoResourceAttr("qbee_rauc.test", "node"),
				),
			},
			// Update to a different template
			{
				Config: providerConfig + `
resource "qbee_rauc" "test" {
  tag = "terraform:acctest:rauc"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_rauc.test", "tag", "terraform:acctest:rauc"),
					resource.TestCheckNoResourceAttr("qbee_rauc.test", "node"),
				),
			},
			// Import tag
			{
				ResourceName:                         "qbee_rauc.test",
				ImportState:                          true,
				ImportStateId:                        "tag:terraform:acctest:rauc",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tag",
			},
			// Update to be for a node
			{
				Config: providerConfig + `
resource "qbee_rauc" "test" {
  node = "integrationtests"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("qbee_rauc.test", "tag"),
					resource.TestCheckResourceAttr("qbee_rauc.test", "node", "integrationtests"),
				),
			},
			// Import testing
			{
				ResourceName:                         "qbee_rauc.test",
				ImportState:                          true,
				ImportStateId:                        "node:integrationtests",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "node",
			},
		},
	})
}
