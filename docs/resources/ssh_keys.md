---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "qbee_ssh_keys Resource - qbee"
subcategory: ""
description: |-
  SSHKeys adds or removes authorized SSH keys for users.
---

# qbee_ssh_keys (Resource)

SSHKeys adds or removes authorized SSH keys for users.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `extend` (Boolean) If the configuration should extend configuration from the parent nodes of the node the configuration is applied to. If set to false, configuration from parent nodes is ignored.
- `users` (Attributes List) The users to set SSH keys for. (see [below for nested schema](#nestedatt--users))

### Optional

- `node` (String) The node for which to set the configuration. Either tag or node is required.
- `tag` (String) The tag for which to set the configuration. Either tag or node is required.

<a id="nestedatt--users"></a>
### Nested Schema for `users`

Required:

- `keys` (List of String) The SSH keys to set for the user.
- `username` (String) Username of the user for which the SSH keys are set.

## Import

Import is supported using the following syntax:

```shell
# qbee_ssh_keys can be imported by specifying the type (tag or node), followed by a colon,
# and finally either the tag or the node id.

terraform import qbee_ssh_keys.example_tag tag:example-tag
terraform import qbee_ssh_keys.example_node node:example-node-id
```