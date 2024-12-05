resource "qbee_sshkeys" "example_tag" {
  tag    = "example-tag"
  extend = true
  users = [
    {
      username = "testuser"
      keys = [
        "key1",
        "key2"
      ]
    }
  ]
}

resource "qbee_sshkeys" "example_node" {
  node   = "example-node-id"
  extend = true
  users = [
    {
      username = "testuser"
      keys     = ["key1"]
    },
    {
      username = "otheruser"
      keys     = []
    }
  ]
}
