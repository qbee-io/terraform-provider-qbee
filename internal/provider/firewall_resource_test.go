package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFirewallResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "qbee_firewall" "test" {
  node = "integrationtests"
  extend = true

  input = {
    policy = "DROP"
    rules = [
      {
        proto = "tcp"
        target = "ACCEPT"
        src_ip = "192.0.2.0/24"
        dst_port = "22"
      },
      {
        proto = "udp"
        target = "ACCEPT"
        src_ip = "198.51.100.0/24"
        dst_port = "50055"
      },
    ]
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_firewall.test", "node", "integrationtests"),
					resource.TestCheckNoResourceAttr("qbee_firewall.test", "tag"),
					resource.TestCheckResourceAttr("qbee_firewall.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_firewall.test", "input.policy", "DROP"),
					resource.TestCheckResourceAttr("qbee_firewall.test", "input.rules.#", "2"),
					resource.TestCheckResourceAttr("qbee_firewall.test", "input.rules.0.proto", "tcp"),
					resource.TestCheckResourceAttr("qbee_firewall.test", "input.rules.0.target", "ACCEPT"),
					resource.TestCheckResourceAttr("qbee_firewall.test", "input.rules.0.src_ip", "192.0.2.0/24"),
					resource.TestCheckResourceAttr("qbee_firewall.test", "input.rules.0.dst_port", "22"),
					resource.TestCheckResourceAttr("qbee_firewall.test", "input.rules.1.proto", "udp"),
					resource.TestCheckResourceAttr("qbee_firewall.test", "input.rules.1.target", "ACCEPT"),
					resource.TestCheckResourceAttr("qbee_firewall.test", "input.rules.1.src_ip", "198.51.100.0/24"),
					resource.TestCheckResourceAttr("qbee_firewall.test", "input.rules.1.dst_port", "50055"),
				),
			},
			{
				Config: providerConfig + `
resource "qbee_firewall" "test" {
  tag = "terraform:acctest:firewall"
  extend = false

  input = {
    policy = "ACCEPT"
    rules = [
      {
        proto = "tcp"
        target = "DROP"
        src_ip = "0.0.0.0/0"
        dst_port = "22"
      },
    ]
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("qbee_firewall.test", "node"),
					resource.TestCheckResourceAttr("qbee_firewall.test", "tag", "terraform:acctest:firewall"),
					resource.TestCheckResourceAttr("qbee_firewall.test", "extend", "false"),
					resource.TestCheckResourceAttr("qbee_firewall.test", "input.policy", "ACCEPT"),
					resource.TestCheckResourceAttr("qbee_firewall.test", "input.rules.#", "1"),
					resource.TestCheckResourceAttr("qbee_firewall.test", "input.rules.0.proto", "tcp"),
					resource.TestCheckResourceAttr("qbee_firewall.test", "input.rules.0.target", "DROP"),
					resource.TestCheckResourceAttr("qbee_firewall.test", "input.rules.0.src_ip", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("qbee_firewall.test", "input.rules.0.dst_port", "22"),
				),
			},
			// Import testing
			{
				ResourceName:                         "qbee_firewall.test",
				ImportState:                          true,
				ImportStateId:                        "tag:terraform:acctest:firewall",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tag",
			},
		},
	})
}
