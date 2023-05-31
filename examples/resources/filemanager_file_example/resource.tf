# Uploads the file 'local/path/file.txt' to '/root/path/file.txt'.
resource "qbee_filemanager_file" "example" {
  path     = "/root/path"
  name     = "file.txt"
  source   = "local/path/file.txt"
  file_sha = filesha1("local/path/file.txt")
}
