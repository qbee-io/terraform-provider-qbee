resource "qbee_users" "example_tag" {
  tag    = "example-tag"
  extend = true
}

resource "qbee_users" "example_node" {
  node   = "example-node-id"
  extend = true
}
