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
	path = "/acctest/filemanager_directory/testdir"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_filemanager_directory.test", "path", "/acctest/filemanager_directory/testdir"),
					resource.TestCheckResourceAttr("qbee_filemanager_directory.test", "id", "placeholder"),
				),
			},
			// Update and read testing
			{
				Config: providerConfig + `
resource "qbee_filemanager_directory" "test" {
	path = "/acctest/filemanager_directory/testdir-2"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_filemanager_directory.test", "path", "/acctest/filemanager_directory/testdir-2"),
					resource.TestCheckResourceAttr("qbee_filemanager_directory.test", "id", "placeholder"),
				),
			},
			// Import testing
			{
				ResourceName:      "qbee_filemanager_directory.test",
				ImportState:       true,
				ImportStateId:     "/acctest/filemanager_directory/testdir-2",
				ImportStateVerify: true,
			},
		},
	})
}
