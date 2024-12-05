package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMetricsMonitorResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "qbee_metrics_monitor" "test" {
  tag = "terraform:acctest:metricsmonitor"
  extend = true
  metrics = [
    {
      value = "cpu:user"
      threshold = 20
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_metrics_monitor.test", "tag", "terraform:acctest:metricsmonitor"),
					resource.TestCheckNoResourceAttr("qbee_metrics_monitor.test", "node"),
					resource.TestCheckResourceAttr("qbee_metrics_monitor.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_metrics_monitor.test", "metrics.#", "1"),
					resource.TestCheckResourceAttr("qbee_metrics_monitor.test", "metrics.0.value", "cpu:user"),
					resource.TestCheckResourceAttr("qbee_metrics_monitor.test", "metrics.0.threshold", "20"),
				),
			},
			// Update to a different template
			{
				Config: providerConfig + `
resource "qbee_metrics_monitor" "test" {
  tag = "terraform:acctest:metricsmonitor"
  extend = false
  metrics = [
    {
      value = "cpu:user"
      threshold = 30
    },
    {
      value = "filesystem:use"
      threshold = 60.5
      id = "/data"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_metrics_monitor.test", "tag", "terraform:acctest:metricsmonitor"),
					resource.TestCheckNoResourceAttr("qbee_metrics_monitor.test", "node"),
					resource.TestCheckResourceAttr("qbee_metrics_monitor.test", "extend", "false"),
					resource.TestCheckResourceAttr("qbee_metrics_monitor.test", "metrics.#", "2"),
					resource.TestCheckResourceAttr("qbee_metrics_monitor.test", "metrics.0.value", "cpu:user"),
					resource.TestCheckResourceAttr("qbee_metrics_monitor.test", "metrics.0.threshold", "30"),
					resource.TestCheckNoResourceAttr("qbee_metrics_monitor.test", "metrics.0.id"),
					resource.TestCheckResourceAttr("qbee_metrics_monitor.test", "metrics.1.value", "filesystem:use"),
					resource.TestCheckResourceAttr("qbee_metrics_monitor.test", "metrics.1.threshold", "60.5"),
					resource.TestCheckResourceAttr("qbee_metrics_monitor.test", "metrics.1.id", "/data"),
				),
			},
			// Import tag
			{
				ResourceName:                         "qbee_metrics_monitor.test",
				ImportState:                          true,
				ImportStateId:                        "tag:terraform:acctest:metricsmonitor",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tag",
			},
			// Update to be for a node
			{
				Config: providerConfig + `
resource "qbee_metrics_monitor" "test" {
  node = "integrationtests"
  extend = true
  metrics = [
    {
      value = "cpu:user"
      threshold = 30
    },
    {
      value = "filesystem:use"
      threshold = 60.2
      id = "/data"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("qbee_metrics_monitor.test", "tag"),
					resource.TestCheckResourceAttr("qbee_metrics_monitor.test", "node", "integrationtests"),
					resource.TestCheckResourceAttr("qbee_metrics_monitor.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_metrics_monitor.test", "metrics.#", "2"),
					resource.TestCheckResourceAttr("qbee_metrics_monitor.test", "metrics.0.value", "cpu:user"),
					resource.TestCheckResourceAttr("qbee_metrics_monitor.test", "metrics.0.threshold", "30"),
					resource.TestCheckNoResourceAttr("qbee_metrics_monitor.test", "metrics.0.id"),
					resource.TestCheckResourceAttr("qbee_metrics_monitor.test", "metrics.1.value", "filesystem:use"),
					resource.TestCheckResourceAttr("qbee_metrics_monitor.test", "metrics.1.threshold", "60.2"),
					resource.TestCheckResourceAttr("qbee_metrics_monitor.test", "metrics.1.id", "/data"),
				),
			},
			// Import testing
			{
				ResourceName:                         "qbee_metrics_monitor.test",
				ImportState:                          true,
				ImportStateId:                        "node:integrationtests",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "node",
			},
		},
	})
}
