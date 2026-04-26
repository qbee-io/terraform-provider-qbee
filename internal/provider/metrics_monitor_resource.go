package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.qbee.io/client/config"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resourceModelManager                  = &metricsMonitorResourceModel{}
	_ resource.Resource                     = &metricsMonitorResource{}
	_ resource.ResourceWithConfigure        = &metricsMonitorResource{}
	_ resource.ResourceWithConfigValidators = &metricsMonitorResource{}
	_ resource.ResourceWithImportState      = &metricsMonitorResource{}
)

// NewMetricsMonitorResource is a helper function to simplify the provider implementation.
func NewMetricsMonitorResource() resource.Resource {
	return &metricsMonitorResource{
		configurationResource: configurationResource{
			resourceBase: newResourceBase(config.MetricsMonitorBundle),
			modelFactory: func() any {
				return new(metricsMonitorResourceModel)
			},
		},
	}
}

type metricsMonitorResource struct {
	configurationResource
}

// Schema defines the schema for the resource.
func (r *metricsMonitorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "MetricsMonitor configures on-agent metrics monitoring.",
		Attributes: map[string]schema.Attribute{
			"tag": schema.StringAttribute{
				Optional:      true,
				Description:   "The tag for which to set the configuration. Either tag or node is required.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"node": schema.StringAttribute{
				Optional:      true,
				Description:   "The node for which to set the configuration. Either tag or node is required.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"extend": schema.BoolAttribute{
				Required: true,
				Description: "If the configuration should extend configuration from the parent nodes of the node " +
					"the configuration is applied to. If set to false, configuration from parent nodes is ignored.",
			},
			"metrics": schema.ListNestedAttribute{
				Required:    true,
				Description: "List of monitors for individual metrics",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"value": schema.StringAttribute{
							Required:    true,
							Description: "Value of the metric (enum defined in the JSON schema)",
						},
						"threshold": schema.Float64Attribute{
							Required:    true,
							Description: "Threshold above which a warning will be created by the device",
						},
						"id": schema.StringAttribute{
							Optional:    true,
							Description: "ID of the resource (e.g. filesystem mount point)",
						},
					},
				},
			},
		},
	}
}

type metricsMonitorResourceModel struct {
	configurationResourceModel
	Metrics []metricMonitor `tfsdk:"metrics"`
}

type metricMonitor struct {
	Value     types.String  `tfsdk:"value"`
	Threshold types.Float64 `tfsdk:"threshold"`
	Id        types.String  `tfsdk:"id"`
}

func (m metricsMonitorResourceModel) getConfigBundle() config.Bundle {
	return config.MetricsMonitorBundle
}

func (m *metricsMonitorResourceModel) fromBundleData(bundleData config.BundleData) error {
	data := bundleData.MetricMonitor
	if data == nil {
		return fmt.Errorf("metrics_monitor configuration not found in bundle data")
	}

	m.Extend = types.BoolValue(data.Extend)

	for _, metric := range data.Metrics {
		monitor := metricMonitor{
			Value:     types.StringValue(metric.Value),
			Threshold: types.Float64Value(metric.Threshold),
		}

		if metric.ID != "" {
			monitor.Id = types.StringValue(metric.ID)
		}

		m.Metrics = append(m.Metrics, monitor)
	}

	return nil
}

func (m metricsMonitorResourceModel) toBundleData(metadata config.Metadata) any {
	bundleData := config.MetricsMonitor{
		Metadata: metadata,
	}

	if metadata.Reset {
		return bundleData
	}

	for _, metric := range m.Metrics {
		bundleData.Metrics = append(bundleData.Metrics, config.MetricMonitor{
			Value:     metric.Value.ValueString(),
			Threshold: metric.Threshold.ValueFloat64(),
			ID:        metric.Id.ValueString(),
		})
	}

	return bundleData
}
