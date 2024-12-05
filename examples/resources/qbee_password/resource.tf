resource "qbee_password" "example_tag" {
  tag    = "example-tag"
  extend = true
}

resource "qbee_password" "example_node" {
  node   = "example-node-id"
  extend = true
}
