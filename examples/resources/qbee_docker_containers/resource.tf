resource "qbee_docker_containers" "example_tag" {
  tag    = "example-tag"
  extend = true
  containers = [
    {
      name          = "container-b"
      image         = "debian:latest"
      docker_args   = "-v /path/to/data-volume:/data --hostname my-hostname"
      env_file      = "/tmp/env"
      command       = "echo"
      pre_condition = "true"
    }
  ]
  registry_auths = [
    {
      server   = "registry.example.com"
      username = "user"
      password = "password"
    }
  ]
}

resource "qbee_docker_containers" "example_node" {
  node   = "example_node"
  extend = true
  containers = [
    {
      name  = "container-a"
      image = "debian:stable"
    }
  ]
}
