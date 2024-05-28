terraform {
  required_providers {
    qbee = {
      source = "qbee.io/terraform"
    }
  }
}

provider "qbee" {
}

resource "qbee_filemanager_directory" "filedist_demo" {
  path = "/filedist-demo"
}

resource "qbee_filemanager_file" "filedist_demo_1" {
  path      = "${qbee_filemanager_directory.filedist_demo.path}/file1.txt"
  sourcefile  = "files/file1.txt"
  file_sha256 = filesha256("${path.module}/files/file1.txt")
}

resource "qbee_filemanager_file" "filedist_demo_2" {
  path      = "${qbee_filemanager_directory.filedist_demo.path}/file2.json"
  sourcefile  = "files/file2.json"
  file_sha256 = filesha256("${path.module}/files/file2.json")
}

resource "qbee_tag_filedistribution" "outofthebox_releasegroup_beta" {
  tag    = "terraform:demo"
  extend = true
  files = [
    {
      command = "echo $(date -u) > /tmp/last-updated"
      templates = [
        {
          source      = qbee_filemanager_file.filedist_demo_1.path,
          destination = "/tmp/demonstration/file1.txt",
          is_template = false
        },
        {
          source      = qbee_filemanager_file.filedist_demo_2.path,
          destination = "/tmp/demonstration/file2.json",
          is_template = true
        }
      ]
      parameters = [
        {
          key   = "value_from_parameters",
          value = "this is a value"
        }
      ]
    }
  ]
}
