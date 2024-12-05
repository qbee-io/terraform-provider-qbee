resource "qbee_proc_watch" "example_tag" {
  tag    = "example-tag"
  extend = true
}

resource "qbee_proc_watch" "example_node" {
  node   = "example-node-id"
  extend = true
}
