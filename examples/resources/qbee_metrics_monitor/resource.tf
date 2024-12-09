resource "qbee_metrics_monitor" "example_tag" {
  tag    = "example-tag"
  extend = true
}

resource "qbee_metrics_monitor" "example_node" {
  node   = "example-node-id"
  extend = true
}
