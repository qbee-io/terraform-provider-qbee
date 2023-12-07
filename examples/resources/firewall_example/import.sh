# Firewall can be imported by specifying the type (tag or node), followed by a colon, and
# finally either the tag or the node id.
terraform import qbee_firewall.example tag:example_tag
terraform import qbee_firewall.example node:example_node
