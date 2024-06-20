terraform {
  required_providers {
    qbee = {
      source = "github.com/github.com/lesteenman/terraform-provider-qbee-lesteenman"
    }
  }
}

provider "qbee" {
}

resource "qbee_grouptree_group" "example_group" {
  ancestor = "root"
  title    = "Managed Group"
  id       = "top-level-node"
}

resource "qbee_grouptree_group" "example_nested_group" {
  ancestor = qbee_grouptree_group.example_group.id
  title    = "Nested Group"
  id       = "nested-node"
}
