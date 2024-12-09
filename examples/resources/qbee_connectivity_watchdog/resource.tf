resource "qbee_connectivity_watchdog" "example_tag" {
  tag       = "example-tag"
  extend    = false
  threshold = 5
}

resource "qbee_connectivity_watchdog" "example_node" {
  node      = "example-node-id"
  extend    = true
  threshold = 3
}
