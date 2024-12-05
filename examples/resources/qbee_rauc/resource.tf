resource "qbee_rauc" "example_tag" {
  tag         = "example-tag"
  extend      = true
  rauc_bundle = "/path/to/bundle.raucb"
}

resource "qbee_rauc" "example_node" {
  node        = "example-node-id"
  extend      = true
  rauc_bundle = "/path/to/bundle.raucb"
}
