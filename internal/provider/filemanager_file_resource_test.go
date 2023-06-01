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
	parent = "/acctest/filemanager_file/"
	sourcefile = "testfiles/file1.txt"
	file_hash = filesha1("testfiles/file1.txt")
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "path", "/acctest/filemanager_file/file1.txt"),
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "parent", "/acctest/filemanager_file/"),
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "name", "file1.txt"),
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "id", "placeholder"),
				),
			},
			// Update test
			{
				Config: providerConfig + `
resource "qbee_filemanager_file" "test" {
	parent = "/acctest/filemanager_file/"
	name = "alt_filename.txt"
	sourcefile = "testfiles/file2.txt"
	file_hash = filesha1("testfiles/file2.txt")
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "path", "/acctest/filemanager_file/alt_filename.txt"),
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "parent", "/acctest/filemanager_file/"),
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "name", "alt_filename.txt"),
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "id", "placeholder"),
				),
			},
			// Import testing
			{
				ResourceName:            "qbee_filemanager_file.test",
				ImportState:             true,
				ImportStateId:           "/acctest/filemanager_file/alt_filename.txt",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"file_hash", "sourcefile"},
			},
		},
	})
}
