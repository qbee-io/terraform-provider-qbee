# Creates a firewall configuration for the tag 'example_tag'
resource "qbee_filemanager_file" "example" {
  tag = "example_tag"

  input = {
    policy = "DROP"
    rules = [
      {
        proto   = "TCP"
        target  = "ACCEPT"
        srcIp   = "192.0.2.0/24"
        dstPort = "22"
      },
      {
        proto   = "TCP"
        target  = "DROP"
        srcIp   = "0.0.0.0/0"
        dstPort = "22"
      },
      {
        proto   = "UDP"
        target  = "ACCEPT"
        srcIp   = "198.51.100.0/24"
        dstPort = "50055"
      },
    ]
  }
}
