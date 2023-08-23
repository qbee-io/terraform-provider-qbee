terraform {
  required_providers {
    qbee = {
      source = "lesteenman/qbee"
    }
  }
}

provider "qbee" {
}

resource "qbee_filemanager_directory" "filedist_demo" {
  parent = "/"
  name = "filedist-demo"
}

resource "qbee_filemanager_file" "filedist_demo_1" {
  parent = qbee_filemanager_directory.filedist_demo.path
  source = 'files/file1.txt'
  file_hash  = filesha1("files/file1.txt")
}

resource "qbee_filemanager_file" "filedist_demo_2" {
  parent = qbee_filemanager_directory.filedist_demo.path
  source = 'files/file2.txt'
  file_hash  = filesha1("files/file2.txt")
}

resource "qbee_tag_filedistribution" "outofthebox_releasegroup_beta" {
  tag = "terraform:demo"
  extend = true
  files = [
    {
      command = "echo $(date -u) > /tmp/last-updated"
      templates = [
        {
          source = qbee_filemanager_file.filedist_demo_1.path,
          destination = "/tmp/demonstration/file1.txt",
          is_template = false
        },
        {
          source = qbee_filemanager_file.filedist_demo_2.path,
          destination = "/tmp/demonstration/file2.txt",
          is_template = true
        }
      ]
      parameters = [
        {
          key = "value_from_parameters",
          value = "this is a value"
        }
      ]
    }
  ]
}
