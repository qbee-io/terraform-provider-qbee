package provider

import (
	"regexp"
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
	path = "/acctest/filemanager_file/file1.txt"
	sourcefile = "testfiles/file1.txt"
	file_sha256 = filesha256("testfiles/file1.txt")
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "path", "/acctest/filemanager_file/file1.txt"),
					resource.TestMatchResourceAttr("qbee_filemanager_file.test", "file_sha256", regexp.MustCompile("^[a-fA-F0-9]{64}$")),
				),
			},
			// Update test
			{
				Config: providerConfig + `
resource "qbee_filemanager_file" "test" {
	path = "/acctest/filemanager_file/alt_filename.txt"
	sourcefile = "testfiles/file2.txt"
	file_sha256 = filesha256("testfiles/file2.txt")
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_filemanager_file.test", "path", "/acctest/filemanager_file/alt_filename.txt"),
					resource.TestMatchResourceAttr("qbee_filemanager_file.test", "file_sha256", regexp.MustCompile("^[a-fA-F0-9]{64}$")),
				),
			},
			// Import testing
			{
				ResourceName:                         "qbee_filemanager_file.test",
				ImportState:                          true,
				ImportStateId:                        "/acctest/filemanager_file/alt_filename.txt",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "path",
				ImportStateVerifyIgnore:              []string{"file_hash", "sourcefile"},
			},
		},
	})
}
