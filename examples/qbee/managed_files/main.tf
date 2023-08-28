terraform {
  required_providers {
    qbee = {
      source = "lesteenman/qbee"
    }
  }
}

provider "qbee" {
}

resource "qbee_filemanager_directory" "demo_dir" {
  parent = "/terraform-demo/"
  name   = "toplevel"
}

resource "qbee_filemanager_directory" "nested_dir" {
  parent = qbee_filemanager_directory.demo_dir.path
  name   = "nested"
}

resource "qbee_filemanager_file" "example_file" {
  parent      = qbee_filemanager_directory.demo_dir.path
  sourcefile  = "files/file1.txt"
  name        = "test-file.txt"
  file_sha256 = filesha256("files/file1.txt")
}
