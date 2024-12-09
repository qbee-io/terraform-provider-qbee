resource "qbee_users" "example_tag" {
  tag    = "example-tag"
  extend = true
  users = [
    {
      username = "test-user-1"
      action   = "add"
    }
  ]
}

resource "qbee_users" "example_node" {
  node   = "example-node-id"
  extend = true
  users = [
    {
      username = "test-user-2"
      action   = "add"
    },
    {
      username = "default-user"
      action   = "remove"
    }
  ]
}
