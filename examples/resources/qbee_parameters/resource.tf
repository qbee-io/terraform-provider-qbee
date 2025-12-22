resource "qbee_parameters" "example_tag" {
  tag    = "example_tag"
  extend = true
  parameters = [
    {
      key   = "parameter-key-1"
      value = "parameter-value-1"
    }
  ]
}

resource "qbee_parameters" "example_with_secrets" {
  tag    = "example_tag"
  extend = true
  secrets = [
    {
      key              = "secret-key"
      value_wo         = "secret-value"
      value_wo_version = "9635d15d-0a2e-ea5b-7bd1-9837802d5fe4"
    }
  ]
}

resource "qbee_filedistribution" "example_node" {
  node   = "example_node"
  extend = true
  parameters = [
    {
      key   = "parameter-key-1"
      value = "$(parameter-value-1)"
    },
    {
      key   = "parameter-key-2"
      value = "$(parameter-value-2)"
    }
  ]
}
