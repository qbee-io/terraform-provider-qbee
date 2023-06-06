terraform {
  required_providers {
    qbee = {
      source = "lesteenman/qbee"
    }
  }
}

provider "qbee" {
}

resource "qbee_grouptree_group" "example_group" {
  parent = "root"
  title = "Managed Group"
}

resource "qbee_grouptree_group" "example_nested_group" {
  parent = qbee_grouptree_group.example_group.id
  title = "Nested Group"
  node_id = "fixed-node-id"
}
