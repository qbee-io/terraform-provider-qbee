resource "qbee_settings" "example_tag" {
  tag    = "example-tag"
  extend = true
}

resource "qbee_settings" "example_node" {
  node   = "example-node-id"
  extend = true
}
