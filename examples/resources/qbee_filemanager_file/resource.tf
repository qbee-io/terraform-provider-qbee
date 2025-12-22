# Uploads the file 'local/path/file.txt' to '/root/path/file.txt'.
resource "qbee_filemanager_file" "example" {
  path        = "/root/path/file.txt"
  sourcefile  = "/tmp/example.txt"
  file_sha256 = filesha256("local/path/file.txt")
}
