resource "qbee_process_watch" "example_tag" {
  tag    = "example-tag"
  extend = true
}

resource "qbee_process_watch" "example_node" {
  node   = "example-node-id"
  extend = true
}
