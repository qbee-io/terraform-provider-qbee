resource "qbee_package_management" "example_tag" {
  tag          = "example-tag"
  extend       = true
  full_upgrade = true
}

resource "qbee_package_management" "example_node" {
  node          = "example-node-id"
  extend        = true
  pre_condition = "true"
  reboot_mode   = "never"
  packages = [
    {
      "name" : "vim",
      "version" : "9.1",
    }
  ]
}
