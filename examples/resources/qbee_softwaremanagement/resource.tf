resource "qbee_softwaremanagement" "example" {
  node = "root"
  # or 'tag = "tagname"'
  extend = true

  items = [
    {
      package       = "mysql",
      service_name  = "service",
      pre_condition = "/a/b/c/",
      config_files = [
        {
          template = "/testtemplate"
          location = "/testlocation"
        }
      ]
      parameters = [
        {
          key   = "param1"
          value = "value1"
        }
      ]
    }
  ]
}
