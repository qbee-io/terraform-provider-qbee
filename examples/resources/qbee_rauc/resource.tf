resource "qbee_rauc" "example_tag" {
  tag    = "example-tag"
  extend = true
}

resource "qbee_rauc" "example_node" {
  node   = "example-node-id"
  extend = true
}