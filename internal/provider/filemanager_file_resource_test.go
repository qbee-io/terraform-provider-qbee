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
	path = "/acctest/filemanager_file/file.txt"
	sourcefile = "testfiles/file1.txt"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "path", "/acctest/filemanager_file/file.txt"),
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "id", "placeholder"),
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "file_sha256", "09431664ec3b83745fea743e5952d5f3e51d0440f11f4953316d7b755e4e97e3"),
				),
			},
			// Different source test
			{
				Config: providerConfig + `
resource "qbee_filemanager_file" "test" {
	path = "/acctest/filemanager_file/file.txt"
	sourcefile = "testfiles/file2.txt"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "path", "/acctest/filemanager_file/file.txt"),
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "id", "placeholder"),
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "file_sha256", "3a47d8e39da04a37ff0945268f70863f8c01d7f8fcb1699b1dbeb84037caf359"),
				),
			},
			// Rename test
			{
				Config: providerConfig + `
resource "qbee_filemanager_file" "test" {
	path = "/acctest/filemanager_file/alt_filename.txt"
	sourcefile = "testfiles/file2.txt"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "path", "/acctest/filemanager_file/alt_filename.txt"),
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "id", "placeholder"),
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "file_sha256", "3a47d8e39da04a37ff0945268f70863f8c01d7f8fcb1699b1dbeb84037caf359"),
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
