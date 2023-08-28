package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGrouptreeGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "qbee_grouptree_group" "test" {
	id = "group-under-tf-test"
	ancestor = "integrationtests"
	title = "Testing Terraform"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_grouptree_group.test", "id", "group-under-tf-test"),
					resource.TestCheckResourceAttr("qbee_grouptree_group.test", "ancestor", "integrationtests"),
					resource.TestCheckResourceAttr("qbee_grouptree_group.test", "title", "Testing Terraform"),
				),
			},
			// Rename test
			{
				Config: providerConfig + `
resource "qbee_grouptree_group" "test" {
	id = "group-under-tf-test"
	ancestor = "integrationtests"
	title = "Testing Terraform - Path 2"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_grouptree_group.test", "id", "group-under-tf-test"),
					resource.TestCheckResourceAttr("qbee_grouptree_group.test", "ancestor", "integrationtests"),
					resource.TestCheckResourceAttr("qbee_grouptree_group.test", "title", "Testing Terraform - Path 2"),
				),
			},
			// Move test
			{
				Config: providerConfig + `
resource "qbee_grouptree_group" "test" {
	id = "group-under-tf-test"
	ancestor = "root"
	title = "Testing Terraform - Path 2"
}
			`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_grouptree_group.test", "id", "group-under-tf-test"),
					resource.TestCheckResourceAttr("qbee_grouptree_group.test", "ancestor", "root"),
					resource.TestCheckResourceAttr("qbee_grouptree_group.test", "title", "Testing Terraform - Path 2"),
				),
			},
			// Double update test
			{
				Config: providerConfig + `
resource "qbee_grouptree_group" "test" {
	id = "group-under-tf-test"
	ancestor = "integrationtests"
	title = "Testing Terraform"
}
			`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_grouptree_group.test", "id", "group-under-tf-test"),
					resource.TestCheckResourceAttr("qbee_grouptree_group.test", "ancestor", "integrationtests"),
					resource.TestCheckResourceAttr("qbee_grouptree_group.test", "title", "Testing Terraform"),
				),
			},
			// Import testing
			{
				ResourceName:      "qbee_grouptree_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
