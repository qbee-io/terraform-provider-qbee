resource "qbee_sshkeys" "example_tag" {
  tag    = "example-tag"
  extend = true
}

resource "qbee_sshkeys" "example_node" {
  node   = "example-node-id"
  extend = true
}
