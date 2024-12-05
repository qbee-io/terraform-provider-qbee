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
  extend = true
  rauc_bundle = "/path/to/bundle.raucb"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_rauc.test", "tag", "terraform:acctest:rauc"),
					resource.TestCheckNoResourceAttr("qbee_rauc.test", "node"),
					resource.TestCheckResourceAttr("qbee_rauc.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_rauc.test", "rauc_bundle", "/path/to/bundle.raucb"),
					resource.TestCheckNoResourceAttr("qbee_rauc.test", "pre_condition"),
				),
			},
			// Update to a different template
			{
				Config: providerConfig + `
resource "qbee_rauc" "test" {
  tag = "terraform:acctest:rauc"
  extend = false
  rauc_bundle = "/path/to/bundle.raucb"
  pre_condition = "true"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_rauc.test", "tag", "terraform:acctest:rauc"),
					resource.TestCheckNoResourceAttr("qbee_rauc.test", "node"),
					resource.TestCheckResourceAttr("qbee_rauc.test", "extend", "false"),
					resource.TestCheckResourceAttr("qbee_rauc.test", "rauc_bundle", "/path/to/bundle.raucb"),
					resource.TestCheckResourceAttr("qbee_rauc.test", "pre_condition", "true"),
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
  extend = true
  rauc_bundle = "/path/to/bundle.raucb"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("qbee_rauc.test", "tag"),
					resource.TestCheckResourceAttr("qbee_rauc.test", "node", "integrationtests"),
					resource.TestCheckResourceAttr("qbee_rauc.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_rauc.test", "rauc_bundle", "/path/to/bundle.raucb"),
					resource.TestCheckNoResourceAttr("qbee_rauc.test", "pre_condition"),
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
