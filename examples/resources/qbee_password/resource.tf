resource "qbee_password" "example_tag" {
  tag    = "example-tag"
  extend = true
  users = [
    {
      username      = "testuser"
      password_hash = "$5$XyfC.GiB.I8hP8cT$eyBg53DYYuWG5iAdd1Lm8T2rO/tsq0As2jbkK1lTi3D"
    }
  ]
}

resource "qbee_password" "example_node" {
  node   = "example-node-id"
  extend = true
  users = [
    {
      username      = "testuser"
      password_hash = "$5$rxiCYTuoljJlNdvd$sD00V.1VO9ePdFkogkTos6mSzQuuZLkjLXxyYAkfjSA"
    },
    {
      username      = "seconduser"
      password_hash = "$5$C1XMOaIfW1niwc1n$qezUJc1c8UPVQwHkyD7BvF5JLQU8dZ0r6uQ4X4e8IbB"
    }
  ]
}
