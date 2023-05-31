package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFilemanagerFileResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "qbee_filemanager_file" "test" {
	parent = "/acctest/"
	sourcefile = "testfiles/file1.txt"
	file_hash = filesha1("testfiles/file1.txt")
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "id", "/acctest/file1.txt"),
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "path", "/acctest/file1.txt"),
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "parent", "/acctest/"),
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "name", "file1.txt"),
				),
			},
			// Update filename test
			{
				Config: providerConfig + `
resource "qbee_filemanager_file" "test" {
	parent = "/acctest/"
	name = "alt_filename.txt"
	sourcefile = "testfiles/file1.txt"
	file_hash = filesha1("testfiles/file1.txt")
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "id", "/acctest/alt_filename.txt"),
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "path", "/acctest/alt_filename.txt"),
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "parent", "/acctest/"),
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "name", "alt_filename.txt"),
				),
			},
			// Update source file test
			{
				Config: providerConfig + `
resource "qbee_filemanager_file" "test" {
	parent = "/acctest/"
	name = "alt_filename.txt"
	sourcefile = "testfiles/file2.txt"
	file_hash = filesha1("testfiles/file2.txt")
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "id", "/acctest/alt_filename.txt"),
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "path", "/acctest/alt_filename.txt"),
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "parent", "/acctest/"),
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "name", "alt_filename.txt"),
				),
			},
			// Import testing
			{
				ResourceName:      "qbee_filemanager_file.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
