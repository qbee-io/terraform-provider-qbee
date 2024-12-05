---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "qbee_metrics_monitor Resource - qbee"
subcategory: ""
description: |-
  MetricsMonitor configures on-agent metrics monitoring.
---

# qbee_metrics_monitor (Resource)

MetricsMonitor configures on-agent metrics monitoring.

## Example Usage

```terraform
resource "qbee_metrics_monitor" "example_tag" {
  tag    = "example-tag"
  extend = true
}

resource "qbee_metrics_monitor" "example_node" {
  node   = "example-node-id"
  extend = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `extend` (Boolean) If the configuration should extend configuration from the parent nodes of the node the configuration is applied to. If set to false, configuration from parent nodes is ignored.
- `metrics` (Attributes List) List of monitors for individual metrics (see [below for nested schema](#nestedatt--metrics))

### Optional

- `node` (String) The node for which to set the configuration. Either tag or node is required.
- `tag` (String) The tag for which to set the configuration. Either tag or node is required.

<a id="nestedatt--metrics"></a>
### Nested Schema for `metrics`

Required:

- `threshold` (Number) Threshold above which a warning will be created by the device
- `value` (String) Value of the metric (enum defined in the JSON schema)

Optional:

- `id` (String) ID of the resource (e.g. filesystem mount point)

## Import

Import is supported using the following syntax:

```shell
# qbee_metrics_monitor can be imported by specifying the type (tag or node), followed by a colon,
# and finally either the tag or the node id.

terraform import qbee_metrics_monitor.example_tag tag:example-tag
terraform import qbee_metrics_monitor.example_node node:example-node-id
```