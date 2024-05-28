terraform {
  required_providers {
    qbee = {
      source = "qbee.io/terraform"
    }
  }
}

provider "qbee" {
}

resource "qbee_filemanager_directory" "demo_dir" {
  path = "/terraform-demo"
}

resource "qbee_filemanager_directory" "nested_dir" {
  path = "${qbee_filemanager_directory.demo_dir.path}/nested"
}

resource "qbee_filemanager_file" "example_file" {
  path        = "${qbee_filemanager_directory.nested_dir.path}/test-file.txt"
  sourcefile  = "files/file1.txt"
  file_sha256 = filesha256("files/file1.txt")
}
