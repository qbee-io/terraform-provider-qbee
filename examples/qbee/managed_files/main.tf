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
  parent     = qbee_filemanager_directory.demo_dir.path
  sourcefile = "files/file1.txt"
  name = "test-file.txt"
  file_hash  = filesha1("files/file1.txt")
}

#resource "qbee_filemanager_file" "file_with_automatic_name" {
#  parent   = qbee_filemanager_directory.demo_dir.path
#  sourcefile = "files/file2.txt"
#  file_hash = filesha1("files/file2.txt")
#}
