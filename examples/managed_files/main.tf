terraform {
  required_providers {
    qbee = {
      source = "lesteenman/qbee"
    }
  }
}

provider "qbee" {
}

resource "qbee_filemanager_directory" "top_level_dir" {
  parent = "/"
  name   = "toplevel"
}

resource "qbee_filemanager_directory" "nested_dir" {
  parent = qbee_filemanager_directory.top_level_dir.path
  name   = "nested"
}

resource "qbee_filemanager_file" "example_file" {
  parent     = "/"
  sourcefile = "files/file1.txt"
  file_hash  = filesha1("files/file1.txt")
}

resource "qbee_filemanager_file" "file_with_automatic_name" {
  parent   = "/"
  sourcefile = "files/file1.txt"
  file_hash = filesha1("files/file1.txt")
}

resource "qbee_filemanager_file" "file_in_managed_dir" {
  path = qbee_filemanager_directory.nested_dir.path
  name = "other-file.txt"
  sourcefile = "files/file.txt"
}
