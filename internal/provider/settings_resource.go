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
	_ resource.Resource                     = &settingsResource{}
	_ resource.ResourceWithConfigure        = &settingsResource{}
	_ resource.ResourceWithConfigValidators = &settingsResource{}
	_ resource.ResourceWithImportState      = &settingsResource{}
)

const (
	errorImportingSettings = "error importing settings resource"
	errorWritingSettings   = "error writing settings resource"
	errorReadingSettings   = "error reading settings resource"
	errorDeletingSettings  = "error deleting settings resource"
)

// NewSettingsResource is a helper function to simplify the provider implementation.
func NewSettingsResource() resource.Resource {
	return &settingsResource{}
}

type settingsResource struct {
	client *client.Client
}

type settingsResourceModel struct {
	Node              types.String `tfsdk:"node"`
	Tag               types.String `tfsdk:"tag"`
	Extend            types.Bool   `tfsdk:"extend"`
	Metrics           types.Bool   `tfsdk:"metrics"`
	Reports           types.Bool   `tfsdk:"reports"`
	RemoteConsole     types.Bool   `tfsdk:"remote_console"`
	SoftwareInventory types.Bool   `tfsdk:"software_inventory"`
	ProcessInventory  types.Bool   `tfsdk:"process_inventory"`
	AgentInterval     types.Int64  `tfsdk:"agent_interval"`
}

func (m settingsResourceModel) typeAndIdentifier() (config.EntityType, string) {
	return typeAndIdentifier(m.Tag, m.Node)
}

// Metadata returns the resource type name.
func (r *settingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_settings"
}

// Configure adds the provider configured client to the resource.
func (r *settingsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

// Schema defines the schema for the resource.
func (r *settingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Settings defines agent settings.",
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
			"metrics": schema.BoolAttribute{
				Required:    true,
				Description: "Metrics collection enabled.",
			},
			"reports": schema.BoolAttribute{
				Required:    true,
				Description: "Reports collection enabled.",
			},
			"remote_console": schema.BoolAttribute{
				Required:    true,
				Description: "RemoteConsole access enabled.",
			},
			"software_inventory": schema.BoolAttribute{
				Required:    true,
				Description: "SoftwareInventory collection enabled.",
			},
			"process_inventory": schema.BoolAttribute{
				Required:    true,
				Description: "ProcessInventory collection enabled.",
			},
			"agent_interval": schema.Int64Attribute{
				Required:    true,
				Description: "AgentInterval defines how often agent reports back to the device hub (in minutes).",
			},
		},
	}
}

func (r *settingsResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("tag"),
			path.MatchRoot("node"),
		),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *settingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan settingsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeSettings(ctx, plan)
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
func (r *settingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from the plan
	var plan settingsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeSettings(ctx, plan)
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
func (r *settingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state *settingsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configType, identifier := state.typeAndIdentifier()

	// Read the real status
	activeConfig, err := r.client.GetActiveConfig(ctx, configType, identifier, config.EntityConfigScopeOwn)
	if err != nil {
		resp.Diagnostics.AddError(errorReadingSettings,
			"error reading the active configuration: "+err.Error())

		return
	}

	// Update the current state
	currentSettings := activeConfig.BundleData.Settings
	if currentSettings == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Extend = types.BoolValue(currentSettings.Extend)
	state.Metrics = types.BoolValue(currentSettings.Metrics)
	state.Reports = types.BoolValue(currentSettings.Reports)
	state.RemoteConsole = types.BoolValue(currentSettings.RemoteConsole)
	state.SoftwareInventory = types.BoolValue(currentSettings.SoftwareInventory)
	state.ProcessInventory = types.BoolValue(currentSettings.ProcessInventory)
	state.AgentInterval = types.Int64Value(int64(currentSettings.AgentInterval))

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *settingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state settingsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
	configType, identifier := state.typeAndIdentifier()
	tflog.Info(ctx, fmt.Sprintf("Deleting settings for %v %v", configType, identifier))

	content := config.Settings{
		Metadata: config.Metadata{
			Reset:   true,
			Version: "v1",
		},
	}

	changeRequest, err := createChangeRequest(config.SettingsBundle, content, configType, identifier)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingSettings,
			err.Error(),
		)
		return
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			errorDeletingSettings,
			err.Error(),
		)
		return
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create settings")
	if err != nil {
		resp.Diagnostics.AddError(errorDeletingSettings,
			"error creating a commit to delete the settings resource: "+err.Error(),
		)

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			resp.Diagnostics.AddError(
				errorDeletingSettings,
				"error deleting uncommitted settings changes: "+err.Error(),
			)
		}

		return
	}
}

// ImportState imports the resource state from the Terraform state.
func (r *settingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	configType, identifier, found := strings.Cut(req.ID, ":")
	if !found || configType == "" || identifier == "" {
		resp.Diagnostics.AddError(
			errorImportingSettings,
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
			errorImportingSettings,
			fmt.Sprintf("Import type must be either 'node' or 'tag'. Got: %q", configType),
		)
		return
	}
}

func (r *settingsResource) writeSettings(ctx context.Context, plan settingsResourceModel) diag.Diagnostics {
	configType, identifier := plan.typeAndIdentifier()
	extend := plan.Extend.ValueBool()

	// Create the resource
	tflog.Info(ctx, fmt.Sprintf("Creating settings for %v %v", configType, identifier))

	content := config.Settings{
		Metadata: config.Metadata{
			Enabled: true,
			Extend:  extend,
			Version: "v1",
		},
		Metrics:           plan.Metrics.ValueBool(),
		Reports:           plan.Reports.ValueBool(),
		RemoteConsole:     plan.RemoteConsole.ValueBool(),
		SoftwareInventory: plan.SoftwareInventory.ValueBool(),
		ProcessInventory:  plan.ProcessInventory.ValueBool(),
		AgentInterval:     int(plan.AgentInterval.ValueInt64()),
	}

	changeRequest, err := createChangeRequest(config.SettingsBundle, content, configType, identifier)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingSettings,
				err.Error(),
			),
		}
	}

	change, err := r.client.CreateConfigurationChange(ctx, changeRequest)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				errorWritingSettings,
				fmt.Sprintf("Error creating a settings resource with qbee: %v", err),
			),
		}
	}

	_, err = r.client.CommitConfiguration(ctx, "terraform: create settings")
	if err != nil {
		diags := diag.Diagnostics{}

		err = fmt.Errorf("error creating a commit for the settings: %w", err)
		diags.AddError(errorWritingSettings, err.Error())

		err = r.client.DeleteConfigurationChange(ctx, change.SHA)
		if err != nil {
			diags.AddError(
				errorWritingSettings,
				fmt.Errorf("error deleting uncommitted settings changes: %w", err).Error(),
			)
		}

		return diags
	}

	return nil
}
