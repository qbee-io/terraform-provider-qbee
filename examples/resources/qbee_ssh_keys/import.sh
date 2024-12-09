# qbee_ssh_keys can be imported by specifying the type (tag or node), followed by a colon,
# and finally either the tag or the node id.

terraform import qbee_ssh_keys.example_tag tag:example-tag
terraform import qbee_ssh_keys.example_node node:example-node-id
