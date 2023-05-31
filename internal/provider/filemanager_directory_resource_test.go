package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFilemanagerDirectoryResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "qbee_filemanager_directory" "test" {
	parent = "/"
	name = "testdir"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_filemanager_directory.test", "id", "/testdir/"),
					resource.TestCheckResourceAttr("qbee_filemanager_directory.test", "path", "/testdir/"),
					resource.TestCheckResourceAttr("qbee_filemanager_directory.test", "parent", "/"),
					resource.TestCheckResourceAttr("qbee_filemanager_directory.test", "name", "testdir"),
				),
			},
			// Update and read testing
			{
				Config: providerConfig + `
resource "qbee_filemanager_directory" "test" {
	parent = "/"
	name = "testdir-2"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_filemanager_directory.test", "id", "/testdir-2/"),
					resource.TestCheckResourceAttr("qbee_filemanager_directory.test", "path", "/testdir-2/"),
					resource.TestCheckResourceAttr("qbee_filemanager_directory.test", "parent", "/"),
					resource.TestCheckResourceAttr("qbee_filemanager_directory.test", "name", "testdir-2"),
				),
			},
			// Import testing
			{
				ResourceName:      "qbee_filemanager_directory.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
