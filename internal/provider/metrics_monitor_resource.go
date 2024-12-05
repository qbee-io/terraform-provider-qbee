package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"go.qbee.io/client"
	"go.qbee.io/client/config"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                     = &metricsMonitorResource{}
	_ resource.ResourceWithConfigure        = &metricsMonitorResource{}
	_ resource.ResourceWithConfigValidators = &metricsMonitorResource{}
	_ resource.ResourceWithImportState      = &metricsMonitorResource{}
)

const (
	errorImportingMetricsMonitor = "error importing metrics_monitor resource"
	errorWritingMetricsMonitor   = "error writing metrics_monitor resource"
	errorReadingMetricsMonitor   = "error reading metrics_monitor resource"
	errorDeletingMetricsMonitor  = "error deleting metrics_monitor resource"
)

// NewMetricsMonitorResource is a helper function to simplify the provider implementation.
func NewMetricsMonitorResource() resource.Resource {
	return &metricsMonitorResource{}
}

type metricsMonitorResource struct {
	client *client.Client
}

type metricsMonitorResourceModel struct {
	Node    types.String    `tfsdk:"node"`
	Tag     types.String    `tfsdk:"tag"`
	Extend  types.Bool      `tfsdk:"extend"`
	Metrics []metricMonitor `tfsdk:"metrics"`
}

func (m metricsMonitorResourceModel) typeAndIdentifier() (config.EntityType, string) {
	return typeAndIdentifier(m.Tag, m.Node)
}

type metricMonitor struct {
	Value     types.String  `tfsdk:"value"`
	Threshold types.Float64 `tfsdk:"threshold"`
	Id        types.String  `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *metricsMonitorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metrics_monitor"
}

// Configure adds the provider configured client to the resource.
func (r *metricsMonitorResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
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

func (r *metricsMonitorResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("tag"),
			path.MatchRoot("node"),
		),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *metricsMonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan metricsMonitorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeMetricsMonitor(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map response body to schema and populate Computed attribute values
	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *metricsMonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from the plan
	var plan metricsMonitorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeMetricsMonitor(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map response body to schema and populate Computed attribute values
	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *metricsMonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state *metricsMonitorResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configType, identifier := state.typeAndIdentifier()

	// Read the real status
	activeConfig, err := r.client.GetActiveConfig(ctx, configType, identifier, config.EntityConfigScopeOwn)
	if err != nil {
		resp.Diagnostics.AddError(errorReadingMetricsMonitor,
			"error reading the active configuration: "+err.Error())

		return
	}

	// Update the current state
	currentMetricsMonitor := activeConfig.BundleData.MetricMonitor
	if currentMetricsMonitor == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Extend = types.BoolValue(currentMetricsMonitor.Extend)

	mappedMetrics := make([]metricMonitor, len(currentMetricsMonitor.Metrics))
	for i, metric := range currentMetricsMonitor.Metrics {
		mappedMetrics[i] = metricMonitor{
			Value:     types.StringValue(metric.Value),
			Threshold: types.Float64Value(metric.Threshold),
		}

		if metric.ID != "" {
			mappedMetrics[i].Id = types.StringValue(metric.ID)
		}
	}
	state.Metrics = mappedMetrics

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *metricsMonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state metricsMonitorResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
	configType, identifier := state.typeAndIdentifier()
	tflog.Info(ctx, fmt.Sprintf("Deleting metrics_monitor for %v %v", configType, identifier))

	content := config.MetricsMonitor{
		Metadata: config.Metadata{
			Reset:   true,
			Version: "v1",
		},
	}

	changeRequest, err := createChangeRequest(config.MetricsMonitorBundle, content, configType, identifier)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingMetricsMonitor,
			err.Error(),
		)
		return
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingMetricsMonitor,
			err.Error(),
		)
		return
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create metrics_monitor")
	if err != nil {
		resp.Diagnostics.AddError(errorDeletingMetricsMonitor,
			"error creating a commit to delete the metrics_monitor resource: "+err.Error(),
		)

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			resp.Diagnostics.AddError(
				errorDeletingMetricsMonitor,
				"error deleting uncommitted metrics_monitor changes: "+err.Error(),
			)
		}

		return
	}
}

// ImportState imports the resource state from the Terraform state.
func (r *metricsMonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	configType, identifier, found := strings.Cut(req.ID, ":")
	if !found || configType == "" || identifier == "" {
		resp.Diagnostics.AddError(
			errorImportingMetricsMonitor,
			fmt.Sprintf("Expected import identifier with format: type:identifier. Got: %q", req.ID),
		)
		return

	}

	if configType == "tag" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tag"), identifier)...)
	} else if configType == "node" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("node"), identifier)...)
	} else {
		resp.Diagnostics.AddError(
			errorImportingMetricsMonitor,
			fmt.Sprintf("Import type must be either 'node' or 'tag'. Got: %q", configType),
		)
		return
	}
}

func (r *metricsMonitorResource) writeMetricsMonitor(ctx context.Context, plan metricsMonitorResourceModel) diag.Diagnostics {
	configType, identifier := plan.typeAndIdentifier()
	extend := plan.Extend.ValueBool()

	// Create the resource
	tflog.Info(ctx, fmt.Sprintf("Creating metrics_monitor for %v %v", configType, identifier))

	mappedMetrics := make([]config.MetricMonitor, len(plan.Metrics))
	for i, metric := range plan.Metrics {
		mappedMetrics[i] = config.MetricMonitor{
			Value:     metric.Value.ValueString(),
			Threshold: metric.Threshold.ValueFloat64(),
			ID:        metric.Id.ValueString(),
		}
	}

	content := config.MetricsMonitor{
		Metadata: config.Metadata{
			Enabled: true,
			Extend:  extend,
			Version: "v1",
		},
		Metrics: mappedMetrics,
	}

	changeRequest, err := createChangeRequest(config.MetricsMonitorBundle, content, configType, identifier)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingMetricsMonitor,
				err.Error(),
			),
		}
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingMetricsMonitor,
				fmt.Sprintf("Error creating a metrics_monitor resource with qbee: %v", err),
			),
		}
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create metrics_monitor")
	if err != nil {
		diags := diag.Diagnostics{}

		err = fmt.Errorf("error creating a commit for the metrics_monitor: %w", err)
		diags.AddError(errorWritingMetricsMonitor, err.Error())

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			diags.AddError(
				errorWritingMetricsMonitor,
				fmt.Errorf("error deleting uncommitted metrics_monitor changes: %w", err).Error(),
			)
		}

		return diags
	}

	return nil
}
