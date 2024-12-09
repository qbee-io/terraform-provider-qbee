package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPodmanContainerResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "qbee_podman_containers" "test" {
  tag = "terraform:acctest:dockercontainer"
  extend = true
  containers = [
    {
      name = "container-a"
      image = "debian:stable"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "tag", "terraform:acctest:dockercontainer"),
					resource.TestCheckNoResourceAttr("qbee_podman_containers.test", "node"),
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "containers.#", "1"),
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "containers.0.name", "container-a"),
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "containers.0.image", "debian:stable"),
				),
			},
			// Update to a different template
			{
				Config: providerConfig + `
resource "qbee_podman_containers" "test" {
  tag = "terraform:acctest:dockercontainer"
  extend = false
  containers = [
    {
      name = "container-b"
      image = "debian:latest"
      podman_args = "-v /path/to/data-volume:/data --hostname my-hostname"
      env_file = "/tmp/env"
      command = "echo"
      pre_condition = "true"
    }
  ]
  registry_auths = [
    {
      server = "registry.example.com"
      username = "user"
      password = "password"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "tag", "terraform:acctest:dockercontainer"),
					resource.TestCheckNoResourceAttr("qbee_podman_containers.test", "node"),
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "extend", "false"),
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "containers.#", "1"),
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "containers.0.name", "container-b"),
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "containers.0.image", "debian:latest"),
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "containers.0.podman_args", "-v /path/to/data-volume:/data --hostname my-hostname"),
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "containers.0.env_file", "/tmp/env"),
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "containers.0.command", "echo"),
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "containers.0.pre_condition", "true"),
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "registry_auths.#", "1"),
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "registry_auths.0.server", "registry.example.com"),
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "registry_auths.0.username", "user"),
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "registry_auths.0.password", "password"),
				),
			},
			// Import tag
			{
				ResourceName:                         "qbee_podman_containers.test",
				ImportState:                          true,
				ImportStateId:                        "tag:terraform:acctest:dockercontainer",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tag",
			},
			// Update to be for a node
			{
				Config: providerConfig + `
resource "qbee_podman_containers" "test" {
  node = "integrationtests"
  extend = true
  containers = [
    {
      name = "container-a"
      image = "debian:stable"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("qbee_podman_containers.test", "tag"),
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "node", "integrationtests"),
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "containers.#", "1"),
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "containers.0.name", "container-a"),
					resource.TestCheckResourceAttr("qbee_podman_containers.test", "containers.0.image", "debian:stable"),
				),
			},
			// Import node
			{
				ResourceName:                         "qbee_podman_containers.test",
				ImportState:                          true,
				ImportStateId:                        "node:integrationtests",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "node",
			},
		},
	})
}
