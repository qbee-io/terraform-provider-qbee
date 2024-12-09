# Parameters can be imported by specifying the type (tag or node), followed by a colon, and
# finally either the tag or the node id.

# Note that if the parameters resource contains secrets, these will not be imported and the resource
# itself will have to be recreated on the next apply.

terraform import qbee_parameters.example tag:example_tag
terraform import qbee_parameters.example node:example_node
