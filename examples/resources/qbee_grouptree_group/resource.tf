# Creates a directory "/root/path/dirname".
resource "qbee_grouptree_group" "example" {
  id       = "grouptree-group-id"
  ancestor = "root"
  title    = "Group Title"
}

resource "qbee_grouptree_group" "example_nested" {
  id       = "grouptree-group-id-nested"
  ancestor = qbee_grouptree_group.example.id
  title    = "Nested Group Title"
}
