resource "qbee_node_filedistribution" "example" {
  tag = "example:tag"

  extend = true
  files = [
    {
      pre_condition = "/bin/true"
      command = "date -u > /tmp/last-updated.txt"
      templates = [
        {
          source = "/example/file.txt.template"
          destination = "/target/path.txt"
          is_template = true
        },
        {
          source = "/example/file2.json"
          destination = "/target/path2.json"
          is_template = false
        }
      ]
      parameters = [
        {
          key = "param-key-1"
          value = "param-value-1"
        },
        {
          key = "param-key-2"
          value = "param-value-2"
        }
      ]
    }
  ]
}
