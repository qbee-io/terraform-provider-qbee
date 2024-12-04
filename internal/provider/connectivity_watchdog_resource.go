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
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                     = &connectivityWatchdogResource{}
	_ resource.ResourceWithConfigure        = &connectivityWatchdogResource{}
	_ resource.ResourceWithConfigValidators = &connectivityWatchdogResource{}
	_ resource.ResourceWithImportState      = &connectivityWatchdogResource{}
)

const (
	errorImportingConnectivityWatchdog = "error importing connectivityWatchdog resource"
	errorWritingConnectivityWatchdog   = "error writing connectivityWatchdog resource"
	errorReadingConnectivityWatchdog   = "error reading connectivityWatchdog resource"
	errorDeletingConnectivityWatchdog  = "error deleting connectivityWatchdog resource"
)

// NewConnectivityWatchdogResource is a helper function to simplify the provider implementation.
func NewConnectivityWatchdogResource() resource.Resource {
	return &connectivityWatchdogResource{}
}

type connectivityWatchdogResource struct {
	client *client.Client
}

type connectivityWatchdogResourceModel struct {
	Node      types.String `tfsdk:"node"`
	Tag       types.String `tfsdk:"tag"`
	ID        types.String `tfsdk:"id"`
	Extend    types.Bool   `tfsdk:"extend"`
	Threshold types.Int64  `tfsdk:"threshold"`
}

func (m connectivityWatchdogResourceModel) typeAndIdentifier() (config.EntityType, string) {
	return typeAndIdentifier(m.Tag, m.Node)
}

// Metadata returns the resource type name.
func (r *connectivityWatchdogResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connectivity_watchdog"
}

// Configure adds the provider configured client to the resource.
func (r *connectivityWatchdogResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Schema defines the schema for the resource.
func (r *connectivityWatchdogResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Placeholder ID value",
			},
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
			"threshold": schema.Int64Attribute{
				Required:    true,
				Description: "defines how many consecutive failed pings are allowed before the watchdog triggers a reboot.",
			},
		},
	}
}

func (r *connectivityWatchdogResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("tag"),
			path.MatchRoot("node"),
		),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *connectivityWatchdogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan connectivityWatchdogResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeConnectivityWatchdog(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue("placeholder")

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *connectivityWatchdogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from the plan
	var plan connectivityWatchdogResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeConnectivityWatchdog(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue("placeholder")

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *connectivityWatchdogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state *connectivityWatchdogResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configType, identifier := state.typeAndIdentifier()

	// Read the real status
	activeConfig, err := r.client.GetActiveConfig(ctx, configType, identifier, config.EntityConfigScopeOwn)
	if err != nil {
		resp.Diagnostics.AddError(errorReadingConnectivityWatchdog,
			"error reading the active configuration: "+err.Error())

		return
	}

	// Update the current state
	currentConnectivityWatchdog := activeConfig.BundleData.ConnectivityWatchdog
	if currentConnectivityWatchdog == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue("placeholder")
	state.Extend = types.BoolValue(currentConnectivityWatchdog.Extend)

	threshold, err := strconv.ParseInt(currentConnectivityWatchdog.Threshold, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			errorReadingConnectivityWatchdog,
			"error parsing the threshold value to an integer value: "+err.Error(),
		)
		return
	}

	state.Threshold = types.Int64Value(threshold)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *connectivityWatchdogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state connectivityWatchdogResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
	configType, identifier := state.typeAndIdentifier()
	tflog.Info(ctx, fmt.Sprintf("Deleting connectivityWatchdog for %v %v", configType, identifier))

	content := config.ConnectivityWatchdog{
		Metadata: config.Metadata{
			Reset:   true,
			Version: "v1",
		},
	}

	changeRequest, err := createChangeRequest(config.ConnectivityWatchdogBundle, content, configType, identifier)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingConnectivityWatchdog,
			err.Error(),
		)
		return
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingConnectivityWatchdog,
			err.Error(),
		)
		return
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create connectivityWatchdog_resource")
	if err != nil {
		resp.Diagnostics.AddError(errorDeletingConnectivityWatchdog,
			"error creating a commit to delete the connectivityWatchdog resource: "+err.Error(),
		)

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			resp.Diagnostics.AddError(
				errorDeletingConnectivityWatchdog,
				"error deleting uncommitted connectivityWatchdog changes: "+err.Error(),
			)
		}

		return
	}
}

// ImportState imports the resource state from the Terraform state.
func (r *connectivityWatchdogResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	configType, identifier, found := strings.Cut(req.ID, ":")
	if !found || configType == "" || identifier == "" {
		resp.Diagnostics.AddError(
			errorImportingConnectivityWatchdog,
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
			errorImportingConnectivityWatchdog,
			fmt.Sprintf("Import type must be either 'node' or 'tag'. Got: %q", configType),
		)
		return
	}
}

func (r *connectivityWatchdogResource) writeConnectivityWatchdog(ctx context.Context, plan connectivityWatchdogResourceModel) diag.Diagnostics {
	configType, identifier := plan.typeAndIdentifier()
	extend := plan.Extend.ValueBool()

	// Create the resource
	tflog.Info(ctx, fmt.Sprintf("Creating connectivityWatchdog for %v %v", configType, identifier))

	content := config.ConnectivityWatchdog{
		Metadata: config.Metadata{
			Enabled: true,
			Extend:  extend,
			Version: "v1",
		},
		Threshold: plan.Threshold.String(),
	}

	changeRequest, err := createChangeRequest(config.ConnectivityWatchdogBundle, content, configType, identifier)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingConnectivityWatchdog,
				err.Error(),
			),
		}
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingConnectivityWatchdog,
				fmt.Sprintf("Error creating a connectivityWatchdog resource with qbee: %v", err),
			),
		}
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create connectivityWatchdog_resource")
	if err != nil {
		diags := diag.Diagnostics{}

		err = fmt.Errorf("error creating a commit for the connectivityWatchdog: %w", err)
		diags.AddError(errorWritingConnectivityWatchdog, err.Error())

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			diags.AddError(
				errorWritingConnectivityWatchdog,
				fmt.Errorf("error deleting uncommitted connectivityWatchdog changes: %w", err).Error(),
			)
		}

		return diags
	}

	return nil
}
