# Creates a directory "/root/path/dirname".
resource "qbee_filemanager_directory" "example" {
  parent = "/root/path/"
  name   = "dirname"
}
